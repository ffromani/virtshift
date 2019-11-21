# virtshift: a toolkit to help managing your openshift cluster on libvirt

The [openshift installer](https://github.com/openshift/installer) allows to install openshift on [KVM virtual machiness managed by libvirt](https://github.com/openshift/installer/blob/master/docs/dev/libvirt/README.md) for development purposes.
`virtshift` provides tools, which don't fit in openshift itself, to help you managing these clusters.

Using openshift on libvirt virtual machines is *strongly* *discouraged* for everything besides development purposes.

## License

Apache v2

## Host-level setup (CentOS/Fedora)

These steps aren't strictly neede, but if they can make your life significantly easier.

### libvirt-guests service
This service takes care of suspending and resuming VMs when the system stops/starts. This may save some `destroy/create` loops.
Steps:
1. configure the service editing `/etc/sysconfig/libvirt-guests`. The file is heavily commented, however a minimal working configuration is show below
```bash
# cat /etc/sysconfig/libvirt-guests
ON_BOOT=start
ON_SHUTDOWN=suspend
```

2. enable the service proper
```
# systemctl enable libvirt-guests.service
```

### openshift-installer fork

[This repo](https://github.com/fromanirh/installer) contains patches to streamline/enhance the libvirt provider workflow. Patches hosted here will be sent as PR once mature
and deemed ready for the broader `openshift-installer` audience.

## Cluster lifecycle

These steps augment the [openshift installer quickstart](https://github.com/openshift/installer/#quick-start). We assume you are familiar with that doc and following it through.
You may want to keep the [docs about using alternate images](https://github.com/openshift/installer/blob/master/docs/dev/alternative_release_image_sources.md) handy.

1. first, prepare your base configuration`install-config.yaml`

2. create the cluster *MANIFESTS*: 
```bash
$ openshift-install --dir=cluster create manifests
```

3. edit the manifests as you see fit. You most likely want to enable [this workaround](https://github.com/openshift/installer/blob/master/docs/dev/libvirt/README.md#console-doesnt-come-up).
By default, VM are quite small. You may want to give them all the resources your box have.
The `virtshift-tune-vms` scripts helps you with that. By default, it autotunes the VMs to consume all the available resources.
```bash
$ cd cluster && virtshift-tune-vms
```
Last, you may want to increase the default timeout. Even on fairly powerful boxes, the default `30m` is likely too low.

4. create the cluster:
```bash
$ openshift install --dir=cluster create cluster
```

5. once the cluster is up -this may requires few tries unfortunately, the first thing you may want to do is to snapshot the disks, so you have a safe point to fall back:
```bash
$ ./virtshift-make-snapshot-sh snapshot-name "human friendly snapshot description" | tee snap.sh

#!/bin/sh
set -ex

PATH="/var/lib/libvirt/openshift-images"
NAME="snapshot-name"
DESC="human friendly snapshot description"
virsh snapshot-create-as test1-68ngb-worker-0-mqtr7 "${NAME}" "${DESC}" --diskspec vda,file="${PATH}/test1-68ngb/test1-68ngb-worker-0-mqtr7-overlay00.qcow2" --disk-only --atomic
virsh snapshot-create-as test1-68ngb-worker-0-rbpp8 "${NAME}" "${DESC}" --diskspec vda,file="${PATH}/test1-68ngb/test1-68ngb-worker-0-rbpp8-overlay00.qcow2" --disk-only --atomic
virsh snapshot-create-as test1-68ngb-master-0 "${NAME}" "${DESC}" --diskspec vda,file="${PATH}/test1-68ngb/test1-68ngb-master-0-overlay00.qcow2" --disk-only --atomic
# review the script
$ bash -x snap.sh
```

6. your cluster is ready to be used!
