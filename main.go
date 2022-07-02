package main

import (
	"context"
	"github.com/trichner/aether/compute"
	"log"
)

var scopes = []string{"https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/compute.readonly", "https://www.googleapis.com/auth/devstorage.full_control", "https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/devstorage.read_write"}

func main() {

	//TODO
	projectID := "example"
	zone := "europe-west6-a"

	service, err := compute.NewService(projectID, zone)
	if err != nil {
		log.Fatal(err)
	}

	instanceName := "aether-009"
	networkName := instanceName + "-net"

	//TODO
	ownerUserEmail := "thomas.richner@example.com"

	ctx := context.Background()

	log.Printf("creating network %q", networkName)
	if err := service.CreateNetwork(ctx, networkName); err != nil {
		log.Fatal(err)
	}

	log.Printf("creating SSH firewall rule for network %q", networkName)
	if err := service.CreateFirewallRuleAllowTcpIngress(ctx, networkName, "ssh", 22); err != nil {
		log.Fatal(err)
	}

	log.Printf("creating ICMP firewall rule for network %q", networkName)
	if err := service.CreateFirewallRuleAllowIcmpIngress(ctx, networkName); err != nil {
		log.Fatal(err)
	}

	log.Printf("creating service account %q", instanceName)
	serviceAccount, err := service.CreateServiceAccount(ctx, instanceName, instanceName+"'s account")
	if err != nil {
		log.Fatal(err)
	}
	email := serviceAccount.Email

	//email := "aether-006@thomas-playground-as2.iam.gserviceaccount.com"
	log.Printf("assigning editor role to service acount %q", email)
	err = service.AddProjectIamBinding(ctx, compute.NewServiceAccountMember(email), "roles/editor")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("assigning computeViewer role to  %q", ownerUserEmail)
	err = service.AddProjectIamBinding(ctx, compute.NewUserMember(ownerUserEmail), "roles/compute.viewer")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("assigning service account access on %q to  %q", email, ownerUserEmail)
	err = service.AddServiceAccountIamBinding(ctx, email, compute.NewUserMember(ownerUserEmail), "roles/iam.serviceAccountUser")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("creating instance %q", instanceName)
	instance := &compute.InsertInstanceRequest{
		Name:                instanceName,
		SourceImage:         "projects/rocky-linux-cloud/global/images/family/rocky-linux-8",
		MachineType:         "e2-standard-4",
		NetworkName:         networkName,
		ServiceAccountEmail: email,
	}
	if err := service.CreateInstance(ctx, instance); err != nil {
		log.Fatal(err)
	}

	log.Printf("assigning instance admin role on %q to  %q", instanceName, ownerUserEmail)
	err = service.AddComputeInstanceIamBinding(ctx, instanceName, compute.NewUserMember(ownerUserEmail), "roles/compute.instanceAdmin")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("assigning OS admin login role on %q to  %q", instanceName, ownerUserEmail)
	err = service.AddComputeInstanceIamBinding(ctx, instanceName, compute.NewUserMember(ownerUserEmail), "roles/compute.osAdminLogin")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("assigning OS admin login role on %q to  %q", instanceName, ownerUserEmail)
	err = service.AddComputeInstanceIamBinding(ctx, instanceName, compute.NewUserMember(ownerUserEmail), "roles/compute.osAdminLogin")
	if err != nil {
		log.Fatal(err)
	}

	// setup VM
}
