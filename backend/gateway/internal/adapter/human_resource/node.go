package human_resource

import (
	"strings"

	"github.com/google/uuid"
)

const nodePrefix = "gid://integrione/"

func nodeID(resource, id string) string {
	return nodePrefix + resource + "/" + id
}

// NodeResource membaca jenis entitas dari ID global untuk routing Gateway.
func NodeResource(id string) string {
	raw, ok := strings.CutPrefix(id, nodePrefix)
	if !ok {
		return ""
	}
	resource, _, _ := strings.Cut(raw, "/")
	return resource
}

func parseNodeID(id, resource string) (uuid.UUID, error) {
	return uuid.Parse(rawNodeID(id, resource))
}

func rawNodeID(id, resource string) string {
	return strings.TrimPrefix(id, nodePrefix+resource+"/")
}

func optionalRawNodeID(id *string, resource string) *string {
	if id == nil {
		return nil
	}
	raw := rawNodeID(*id, resource)
	return &raw
}
