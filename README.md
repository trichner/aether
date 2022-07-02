## Setup gcloud

```bash
# create a new config and login with your account
# this makes it easy to switch between multiple
gcloud config configurations create aether
gcloud config configurations activate aether
gcloud auth login
```

## Required APIs

```
https://console.cloud.google.com/marketplace/product/google/iam.googleapis.com
https://console.cloud.google.com/apis/library/cloudresourcemanager.googleapis.com
https://console.cloud.google.com/apis/library/iam.googleapis.com
https://console.cloud.google.com/apis/library/compute.googleapis.com
```

## Setup OS Login

```bash
# list your keys and username
gcloud compute os-login describe-profile

# generate a new SSH key (or use an existing ID)
ssh-keygen -t ed25519 -C '<firstname.lastname@example.com>'  -f ~/.ssh/id_ed25519_oslogin

# add the key to OS login
gcloud compute os-login ssh-keys add --key-file=~/.ssh/id_ed25519_oslogin.pub
```

**IMPORTANT:** At this point in time, Rocky Linux 8.5 does not support YubiKeys

- [Troubleshooting OS login](https://cloud.google.com/compute/docs/troubleshooting/troubleshoot-os-login)

## SSH Config

```ssh
Host aether
    # aether ip:
    Hostname tbd
    #FIXME:
    User thomas_richner_example_com
    
    IdentityFile ~/.ssh/id_ed25519_oslogin
    
    ServerAliveInterval 30
    
    # control master
    ControlPath ~/.ssh/socket.d/%r@%h-%p
    ControlMaster auto
    ControlPersist 600
    
    
Host aether.tun
    Hostname 34.65.83.175
    #FIXME:
    User thomas_richner_example_com
    
    IdentityFile ~/.ssh/id_ed25519_oslogin
    
    ServerAliveInterval 30
    
    # control master
    ControlPath ~/.ssh/socket.d/%r@%h-%p-tun
    ControlMaster auto
    ControlPersist 600
    
    # all port forwards, adapt as necessary
    LocalForward 3306  127.0.0.1:3306
    LocalForward 3307  127.0.0.1:3307
    LocalForward 4078  127.0.0.1:4078
    LocalForward 4080  127.0.0.1:4080
    LocalForward 5005  127.0.0.1:5005
    LocalForward 8080  127.0.0.1:8080
    LocalForward 8081  127.0.0.1:8081
    LocalForward 8082  127.0.0.1:8082
    LocalForward 8083  127.0.0.1:8083
    LocalForward 8084  127.0.0.1:8084
    LocalForward 8161  127.0.0.1:8161
    LocalForward 8280  127.0.0.1:8280  
    LocalForward 8380  127.0.0.1:8380  
    LocalForward 8787  127.0.0.1:8787
    LocalForward 8790  127.0.0.1:8790
    LocalForward 8791  127.0.0.1:8791
    LocalForward 8792  127.0.0.1:8792
    LocalForward 8793  127.0.0.1:8793
    LocalForward 9990  127.0.0.1:9990
    LocalForward 9993  127.0.0.1:9993
    LocalForward 9994  127.0.0.1:9994
    LocalForward 9995  127.0.0.1:9995
    LocalForward 9996  127.0.0.1:9996
    LocalForward 18080 127.0.0.1:18080
    LocalForward 18787 127.0.0.1:18787
    LocalForward 28080 127.0.0.1:28080
    LocalForward 23306 127.0.0.1:23306
    LocalForward 33306 127.0.0.1:33306
    LocalForward 19990 127.0.0.1:19990
    LocalForward 49232 127.0.0.1:49232
    LocalForward 61616 127.0.0.1:61616
```

Add user to OS login for a given instance:
https://cloud.google.com/compute/docs/oslogin/set-up-oslogin
https://cloud.google.com/compute/docs/access/managing-access-to-resources#bind-member



## Troubleshooting

```
gcloud --project=aethers compute instances get-iam-policy --zone=europe-west6-a aether-003
```

## Installing Maven 3.8 & Java 17
Some examples require maven & Java in a later version than the one provided by the distributuon (3.5). While java can simply be downloaded via dnf
```
sudo dnf install java-17-openjdk-devel
```
After the installation, make sure to enable the correct java version using
```
sudo update-alternatives --config java_sdk_openjdk
```
Maven requires installation from (download)[https://maven.apache.org/download.cgi]. Make sure you setup the mavenrc file to point to the correct jdk, i.e. 
```
[thomas_richner_example_com@aether-v2 ~]$ cat ~/.mavenrc
export JAVA_HOME=/usr/lib/jvm/java-openjdk
```
