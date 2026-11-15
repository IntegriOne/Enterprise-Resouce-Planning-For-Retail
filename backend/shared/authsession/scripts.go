package authsession

import "github.com/redis/go-redis/v9"

// scriptTouchActivity memperpanjang jendela aktivitas hanya jika kunci aktivitas masih ada dan cocok user/gen/epoch (hanya memodifikasi sess:act:*).
var scriptTouchActivity = redis.NewScript(`
local raw = redis.call('GET', KEYS[1])
if not raw then
  return 0
end

local ok, obj = pcall(cjson.decode, raw)
if not ok or type(obj) ~= 'table' then
  return 0
end

if tostring(obj.user_id) ~= ARGV[1] then
  return 0
end
if tonumber(obj.generation) ~= tonumber(ARGV[2]) then
  return 0
end
if tonumber(obj.auth_epoch) ~= tonumber(ARGV[3]) then
  return 0
end

local now_unix = tonumber(ARGV[4])
local new_idle_exp = tonumber(ARGV[5])
local ttl_sec = tonumber(ARGV[6])
if not ttl_sec or ttl_sec <= 0 then
  return 0
end

local cur_seen = tonumber(obj.last_seen_unix) or 0
if now_unix > cur_seen then
  obj.last_seen_unix = now_unix
end
local cur_idle = tonumber(obj.idle_expires_unix) or 0
if new_idle_exp > cur_idle then
  obj.idle_expires_unix = new_idle_exp
end

redis.call('SET', KEYS[1], cjson.encode(obj), 'EX', ttl_sec, 'XX')
return 1
`)

// scriptWriteSessionPair menyimpan token sesi, aktivitas awal, dan state keamanan user secara monotonik oleh iam_writer.
var scriptWriteSessionPair = redis.NewScript(`
local abs_ttl = tonumber(ARGV[2])
local idle_ttl = tonumber(ARGV[4])
if not abs_ttl or abs_ttl <= 0 or not idle_ttl or idle_ttl <= 0 then
  return redis.error_reply('invalid session TTL')
end

local incoming_epoch = tonumber(ARGV[8]) or 0
local incoming_rev = tonumber(ARGV[9]) or 0
local incoming_rec_epoch = tonumber(ARGV[10]) or 0
local incoming_now = tonumber(ARGV[11]) or 0

local cur_sec_raw = redis.call('GET', KEYS[3])
local sec_obj = {
  user_id = ARGV[5],
  status = ARGV[6],
  locked_until_unix = tonumber(ARGV[7]) or 0,
  auth_epoch = incoming_epoch,
  perm_revision = incoming_rev,
  recovery_epoch = incoming_rec_epoch,
  updated_at_unix = incoming_now
}

if cur_sec_raw then
  local ok, old = pcall(cjson.decode, cur_sec_raw)
  if ok and type(old) == 'table' then
    local old_epoch = tonumber(old.auth_epoch) or 0
    local old_rev = tonumber(old.perm_revision) or 0
    local old_rec = tonumber(old.recovery_epoch) or 0
    if old_epoch > sec_obj.auth_epoch then
      sec_obj.auth_epoch = old_epoch
    end
    if old_rev > sec_obj.perm_revision then
      sec_obj.perm_revision = old_rev
    end
    if old_rec > sec_obj.recovery_epoch then
      sec_obj.recovery_epoch = old_rec
    end
    if (old_epoch > incoming_epoch) or (old_rev > incoming_rev) then
      sec_obj.status = old.status
      sec_obj.locked_until_unix = tonumber(old.locked_until_unix) or 0
    end
  end
end

redis.call('SET', KEYS[1], ARGV[1], 'EX', abs_ttl)
redis.call('SET', KEYS[2], ARGV[3], 'EX', idle_ttl)
redis.call('SET', KEYS[3], cjson.encode(sec_obj))
return 1
`)

// scriptBumpUserSecurity memutakhirkan user:sec secara monotonik dan menghapus snapshot lama bila diperlukan.
var scriptBumpUserSecurity = redis.NewScript(`
local incoming_epoch = tonumber(ARGV[4]) or 0
local incoming_rev = tonumber(ARGV[5]) or 0
local incoming_rec = tonumber(ARGV[6]) or 0
local incoming_now = tonumber(ARGV[7]) or 0
local invalidate_authz = ARGV[8]

local next_obj = {
  user_id = ARGV[1],
  status = ARGV[2],
  locked_until_unix = tonumber(ARGV[3]) or 0,
  auth_epoch = incoming_epoch,
  perm_revision = incoming_rev,
  recovery_epoch = incoming_rec,
  updated_at_unix = incoming_now
}

local cur_raw = redis.call('GET', KEYS[1])
if cur_raw then
  local ok, old = pcall(cjson.decode, cur_raw)
  if ok and type(old) == 'table' then
    local old_epoch = tonumber(old.auth_epoch) or 0
    local old_rev = tonumber(old.perm_revision) or 0
    local old_rec = tonumber(old.recovery_epoch) or 0
    local old_updated = tonumber(old.updated_at_unix) or 0
    if old_epoch > incoming_epoch then
      next_obj.auth_epoch = old_epoch
      next_obj.status = old.status
      next_obj.locked_until_unix = tonumber(old.locked_until_unix) or 0
    end
    if old_rev > incoming_rev then
      next_obj.perm_revision = old_rev
    end
    if old_rec > incoming_rec then
      next_obj.recovery_epoch = old_rec
    end
    if old_epoch == incoming_epoch and old_rev == incoming_rev and old_updated > incoming_now then
      next_obj.status = old.status
      next_obj.locked_until_unix = tonumber(old.locked_until_unix) or 0
      next_obj.updated_at_unix = old_updated
    end
  end
end

redis.call('SET', KEYS[1], cjson.encode(next_obj))
if invalidate_authz == '1' and KEYS[2] and KEYS[2] ~= '' then
  redis.call('DEL', KEYS[2])
end
return {next_obj.auth_epoch, next_obj.perm_revision}
`)

// scriptSetGate menulis gerbang keamanan global IAM secara monotonik menurut recovery_epoch.
var scriptSetGate = redis.NewScript(`
local open_flag = ARGV[1] == '1'
local incoming_rec = tonumber(ARGV[2]) or 0
local now_unix = tonumber(ARGV[3]) or 0
local source = ARGV[4]

local cur_rec = tonumber(redis.call('GET', KEYS[2]) or '0') or 0
if cur_rec > incoming_rec then
  incoming_rec = cur_rec
end

local gate_obj = {
  open = open_flag,
  recovery_epoch = incoming_rec,
  updated_at_unix = now_unix,
  source = source
}

redis.call('SET', KEYS[1], cjson.encode(gate_obj))
redis.call('SET', KEYS[2], tostring(incoming_rec))
return incoming_rec
`)

// scriptRateLimit menjalankan counter jendela tetap atomik di namespace Gateway.
var scriptRateLimit = redis.NewScript(`
local win = tonumber(ARGV[1])
if not win or win <= 0 then
  return redis.error_reply('invalid rate limit window')
end
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('EXPIRE', KEYS[1], win)
end
local ttl = redis.call('TTL', KEYS[1])
return {count, ttl}
`)
