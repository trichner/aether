package compute

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComputeService_AddServiceAccountIamBinding(t *testing.T) {

	projectID := "example"
	zone := "europe-west6-a"

	service, err := NewService(projectID, zone)
	assert.NoError(t, err)

	ownerUserEmail := "thomas.richner@example.com"
	serviceAccount := "aether-006@example.iam.gserviceaccount.com"

	ctx := context.Background()

	err = service.AddServiceAccountIamBinding(ctx, serviceAccount, NewUserMember(ownerUserEmail), "roles/iam.serviceAccountUser")
	assert.NoError(t, err)
}
