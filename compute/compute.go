package compute

import (
	compute "cloud.google.com/go/compute/apiv1"
	"context"
	"fmt"
	"github.com/trichner/oauthflows"
	resource "google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
	"net/http"
	"strconv"
)

//TODO
const clientCredentialsPath = "./client_secret.json"

type ComputeService struct {
	oauthClient *http.Client
	projectID   string
	zone        string
}

func NewService(project, zone string) (*ComputeService, error) {

	serviceScopes := []string{
		"https://www.googleapis.com/auth/iam",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/compute.readonly",
		"https://www.googleapis.com/auth/devstorage.full_control",
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/devstorage.read_write",
	}

	client, err := oauthflows.NewClient(oauthflows.WithClientSecretsFile(clientCredentialsPath, serviceScopes), oauthflows.WithFileTokenStore())
	if err != nil {
		return nil, err
	}

	return &ComputeService{
		oauthClient: client,
		projectID:   project,
		zone:        zone,
	}, nil
}

type ServiceAccount struct {
	Name  string
	Email string
}

func (c *ComputeService) CreateServiceAccount(ctx context.Context, name, displayName string) (*ServiceAccount, error) {

	service, err := iam.NewService(ctx, option.WithHTTPClient(c.oauthClient))
	if err != nil {
		return nil, fmt.Errorf("iam.NewService: %w", err)
	}

	request := &iam.CreateServiceAccountRequest{
		AccountId: name,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: displayName,
		},
	}
	account, err := service.Projects.ServiceAccounts.Create("projects/"+c.projectID, request).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot create service account: %w", err)
	}

	return &ServiceAccount{
		Name:  account.Name,
		Email: account.Email,
	}, nil
}

type Member string

func NewServiceAccountMember(email string) Member {
	return Member("serviceAccount:" + email)
}

func NewUserMember(email string) Member {
	return Member("user:" + email)
}

func NewGroupMember(email string) Member {
	return Member("group:" + email)
}

func (m *Member) String() string {
	return string(*m)
}

func (c *ComputeService) AddProjectIamBinding(ctx context.Context, member Member, role string) error {

	projectResource := "projects/" + c.projectID

	service, err := resource.NewService(ctx, option.WithHTTPClient(c.oauthClient))
	if err != nil {
		return fmt.Errorf("cannot create resourcemanager: %w", err)
	}

	policy, err := service.Projects.GetIamPolicy(projectResource, &resource.GetIamPolicyRequest{}).Do()
	if err != nil {
		return fmt.Errorf("cannot get policy: %w", err)
	}

	policy.Bindings = append(policy.Bindings, &resource.Binding{
		Condition: nil,
		Members:   []string{member.String()},
		Role:      role,
	})

	policy, err = service.Projects.SetIamPolicy(projectResource, &resource.SetIamPolicyRequest{Policy: policy}).Do()
	if err != nil {
		return fmt.Errorf("cannot set policy: %w", err)
	}

	return nil
}

func (c *ComputeService) AddServiceAccountIamBinding(ctx context.Context, serviceAccountEmail string, member Member, role string) error {

	client, err := iam.NewService(ctx, option.WithHTTPClient(c.oauthClient))
	if err != nil {
		return err
	}

	resourceId := fmt.Sprintf("projects/%s/serviceAccounts/%s", c.projectID, serviceAccountEmail)

	policy, err := client.Projects.ServiceAccounts.GetIamPolicy(resourceId).Do()
	if err != nil {
		return fmt.Errorf("cannot get serviceAccount %q policy: %w", serviceAccountEmail, err)
	}

	policy.Bindings = append(policy.Bindings, &iam.Binding{
		Members: []string{member.String()},
		Role:    role,
	})

	setRequest := &iam.SetIamPolicyRequest{Policy: policy}

	policy, err = client.Projects.ServiceAccounts.SetIamPolicy(resourceId, setRequest).Do()
	if err != nil {
		return fmt.Errorf("cannot update policy for serviceAccount %q: %w", serviceAccountEmail, err)
	}
	return nil
}

func (c *ComputeService) AddComputeInstanceIamBinding(ctx context.Context, instanceName string, member Member, role string) error {

	client, err := compute.NewInstancesRESTClient(ctx, option.WithHTTPClient(c.oauthClient))
	getRequest := &computepb.GetIamPolicyInstanceRequest{
		Project:  c.projectID,
		Resource: instanceName,
		Zone:     c.zone,
	}
	policy, err := client.GetIamPolicy(ctx, getRequest)
	if err != nil {
		return fmt.Errorf("cannot get instance %q policy: %w", instanceName, err)
	}

	policy.Bindings = append(policy.Bindings, &computepb.Binding{
		Members: []string{member.String()},
		Role:    proto.String(role),
	})

	setRequest := &computepb.SetIamPolicyInstanceRequest{
		Project:  c.projectID,
		Resource: instanceName,
		Zone:     c.zone,
		ZoneSetPolicyRequestResource: &computepb.ZoneSetPolicyRequest{
			Policy: policy,
		},
	}

	_, err = client.SetIamPolicy(ctx, setRequest)
	if err != nil {
		return fmt.Errorf("cannot set instance %q policy: %w", instanceName, err)
	}

	return nil
}

//roles/editor

type InsertInstanceRequest struct {
	//Name the name of the instance, e.g. 'my-machine-03'
	Name string

	//SourceImage the image to use, e.g. 'projects/debian-cloud/global/images/family/debian-10'
	SourceImage string

	//MachineType such as 'e2-standard-4'
	MachineType string

	//NetworkName the name of the network to attach the instance to
	NetworkName string

	//ServiceAccountEmail the identity to assign the instance
	ServiceAccountEmail string
}

func (c *ComputeService) CreateInstance(ctx context.Context, instance *InsertInstanceRequest) error {

	machineTypeUri := fmt.Sprintf("zones/%s/machineTypes/%s", c.zone, instance.MachineType)
	diskTypeUri := fmt.Sprintf("projects/%s/zones/%s/diskTypes/pd-balanced", c.projectID, c.zone)

	instanceScopes := []string{
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/logging.write",
		"https://www.googleapis.com/auth/monitoring.write",
		"https://www.googleapis.com/auth/servicecontrol",
		"https://www.googleapis.com/auth/service.management.readonly",
		"https://www.googleapis.com/auth/trace.append",
	}

	req := &computepb.InsertInstanceRequest{
		Project: c.projectID,
		Zone:    c.zone,
		InstanceResource: &computepb.Instance{
			Disks: []*computepb.AttachedDisk{
				{
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskSizeGb:  proto.Int64(64),
						DiskType:    proto.String(diskTypeUri),
						SourceImage: proto.String(instance.SourceImage),
					},
					AutoDelete: proto.Bool(true),
					Boot:       proto.Bool(true),
					Type:       proto.String(computepb.AttachedDisk_PERSISTENT.String()),
					DeviceName: proto.String(instance.Name),
				},
			},
			MachineType: proto.String(machineTypeUri),
			Metadata: &computepb.Metadata{
				Items: []*computepb.Items{{
					Key: proto.String("enable-oslogin"), Value: proto.String("true"),
				},
				},
			},
			Name: proto.String(instance.Name),
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Name: proto.String(instance.NetworkName),
					AccessConfigs: []*computepb.AccessConfig{
						// required to get an ephemeral external IP
						{
							NetworkTier: proto.String(computepb.AccessConfig_PREMIUM.String()),
						}},
				},
			},
			Scheduling: &computepb.Scheduling{
				OnHostMaintenance: proto.String(computepb.Scheduling_MIGRATE.String()),
			},
			ServiceAccounts: []*computepb.ServiceAccount{
				{
					Email:  proto.String(instance.ServiceAccountEmail),
					Scopes: instanceScopes,
				},
			},
			ShieldedInstanceConfig: &computepb.ShieldedInstanceConfig{
				EnableIntegrityMonitoring: proto.Bool(true),
				EnableSecureBoot:          proto.Bool(false),
				EnableVtpm:                proto.Bool(true),
			},
		},
	}

	client, err := compute.NewInstancesRESTClient(ctx, option.WithHTTPClient(c.oauthClient))
	if err != nil {
		return err
	}

	defer client.Close()

	op, err := client.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create instance: %w", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for instance creation: %w", err)
	}
	return nil
}

func (c *ComputeService) CreateNetwork(ctx context.Context, name string) error {

	// https://cloud.google.com/compute/docs/reference/rest/v1/networks/insert
	req := &computepb.InsertNetworkRequest{
		NetworkResource: &computepb.Network{
			AutoCreateSubnetworks: proto.Bool(true),
			Mtu:                   proto.Int32(1460),
			Name:                  proto.String(name),
			RoutingConfig: &computepb.NetworkRoutingConfig{
				RoutingMode: proto.String("REGIONAL"),
			},
		},
		Project: c.projectID,
	}

	client, err := compute.NewNetworksRESTClient(ctx, option.WithHTTPClient(c.oauthClient))
	if err != nil {
		return err
	}

	defer client.Close()

	op, err := client.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create network: %w", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %w", err)
	}
	return nil
}

func (c *ComputeService) CreateFirewallRuleAllowIcmpIngress(ctx context.Context, networkName string) error {

	network := c.getNetworkUrl(networkName)
	ruleName := fmt.Sprintf("%s-allow-icmp", networkName)
	// https://cloud.google.com/compute/docs/reference/rest/v1/firewalls/insert
	req := &computepb.InsertFirewallRequest{
		FirewallResource: &computepb.Firewall{
			Allowed: []*computepb.Allowed{{
				IPProtocol: proto.String("icmp"),
			}},
			Direction:    proto.String("INGRESS"),
			Name:         proto.String(ruleName),
			Network:      proto.String(network),
			Priority:     proto.Int32(65534),
			SourceRanges: []string{"0.0.0.0/0"},
		},
		Project: c.projectID,
	}

	return c.doFirewallInsert(ctx, req)
}

func (c *ComputeService) CreateFirewallRuleAllowTcpIngress(ctx context.Context, networkName string, serviceName string, port int) error {

	network := c.getNetworkUrl(networkName)
	ruleName := fmt.Sprintf("%s-allow-%s", networkName, serviceName)
	// https://cloud.google.com/compute/docs/reference/rest/v1/firewalls/insert
	req := &computepb.InsertFirewallRequest{
		FirewallResource: &computepb.Firewall{
			Allowed: []*computepb.Allowed{{
				IPProtocol: proto.String("tcp"),
				Ports:      []string{strconv.Itoa(port)},
			}},
			Direction:    proto.String("INGRESS"),
			Name:         proto.String(ruleName),
			Network:      proto.String(network),
			Priority:     proto.Int32(65534),
			SourceRanges: []string{"0.0.0.0/0"},
		},
		Project: c.projectID,
	}

	return c.doFirewallInsert(ctx, req)
}

func (c *ComputeService) getNetworkUrl(networkName string) string {
	return fmt.Sprintf("projects/%s/global/networks/%s", c.projectID, networkName)
}

func (c *ComputeService) doFirewallInsert(ctx context.Context, req *computepb.InsertFirewallRequest) error {

	client, err := compute.NewFirewallsRESTClient(ctx, option.WithHTTPClient(c.oauthClient))
	if err != nil {
		return err
	}

	defer client.Close()

	op, err := client.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create network: %w", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %w", err)
	}
	return nil
}
