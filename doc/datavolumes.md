# Data Volumes

## Introduction
Data Volumes(DV) are an abstraction on top of Persistent Volume Claims(PVC) and the Containerized Data Importer(CDI). The DV will monitor and orchestrate the import/upload/clone of the data into the PVC.

Why is this an improvement over simply looking at the state annotation created and managed by CDI? Data Volumes provide a versioned API that other projects like [Kubevirt](https://github.com/kubevirt/kubevirt) can integrate with. This way those projects can rely on an API staying the same for a particular version and have guarantees about what that API will look like. Any changes to the API will result in a new version of the API.

### Status phases
The following statuses are possible.
* 'Blank': No status available.
* Pending: The operation is pending, but has not been scheduled yet.
* WaitForFirstConsumer: The PVC associated with the operation is Pending, and the storage has
  a WaitForFirstConsumer binding mode. PVC [waits for a consumer](waitforfirstconsumer-storage-handling.md) Pod.
* PVCBound: The PVC associated with the operation has been bound.
* Import/Clone/UploadScheduled: The operation (import/clone/upload) has been scheduled.
* Import/Clone/UploadInProgress: The operation (import/clone/upload) is in progress.
* SnapshotForSmartClone/SmartClonePVCInProgress: The Smart-Cloning operation is in progress.
* CSICloneInProgress: The CSI Volume Clone operation is in progress
* CloneFromSnapshotSourceInProgress: Clone from VolumeSnapshot source is in progress
* Paused: A [multi-stage](#multi-stage-import) import is waiting to transfer a new checkpoint.
* Succeeded: The operation has succeeded.
* Failed: The operation has failed.
* Unknown: Unknown status.

## Target Storage/PVC

There are two ways to request storage - by using either the `pvc` or the `storage` section in the DataVolume resource yaml.
Both result in CDI creating a PVC resource, but there are some differences in how they work.

> [!NOTE] 
> We recommend using the storage API as it has more convenience features.

### PVC
The `pvc` type specifies the PersistentVolumeClaim resource that will be created by the CDI. 
All the parameters of pvc have the semantics of PersistentVolumeClaim parameters, 
e.g when the volumeMode is not specified the kubernetes default Filesystem is used. The example shows 
  that a PVC with at least 1Gi of storage and ReadWriteOnce accessMode will be created. 

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: blank-dv-with-pvc
spec:
  source:
    blank: {}
  pvc:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

### Storage
The `storage` type is similar to `pvc` but it allows you to omit some parameters.

> [!TIP]
> With the storage API, CDI computes virtualization-specific defaults for optional
> fields, otherwise falling back to k8's.

If you skip the `volumeMode` parameter, CDI will search for a default value in the
StorageProfile. See also: [storage profile documentation](storageprofile.md).

If the volume mode is set to `fileSystem` (either explicitly with `volumeMode: fileSystem`,
or implicitly as explained in the previous paragraph), CDI will take the file system
overhead into account and request a PVC big enough to fit both an image and the file
system metadata. This logic only applies to the `DataVolume.spec.storage`.

#### Default virtualization storage class
If you skip the `storageClassName` parameter, CDI will prioritize the default
virtualization storage class over k8s' default. You can define your default
virtualization storage class by annotating it with
`storageclass.kubevirt.io/is-default-virt-class` set to `"true"`.

This example shows a request for a PVC with at least 1Gi of storage. Other fields are
left for CDI to fill in.
```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: blank-dv-with-storage
spec:
  source:
    blank: {}
  storage:
    resources:
      requests:
        storage: 1Gi
```
> [!WARNING]
> The detection and automation of storage parameters can vary depending on the `source`.
> For example, cloning a PVC ([PVC source](#pvc-source)) allows for the ommission of storage size,
> otherwise mandatory. Make sure you read the docs for each individual source for more information.

### Block Volume Mode
You can import, clone and upload a disk image to a raw block persistent volume, although  
some CRIs need manual configuration to allow our rootless workload pods to utilize block devices, see [Configure CRI ownership from security context](block_cri_ownership_config.md).  

Block disk image operations are initiated by assigning the value 'Block' to the PVC volumeMode field in the DataVolume yaml.
The following is an example to import disk image to a raw block volume:
```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-import-dv"
spec:
  source:
      http:
         url: "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img" # S3 or GCS
         secretRef: "" # Optional
         certConfigMap: "" # Optional
  storage:
    volumeMode: Block
    resources:
      requests:
        storage: "64Mi"
```

## Source 

### HTTP/S3/GCS/Registry source
DataVolumes are an abstraction on top of the annotations one can put on PVCs to trigger CDI. As such DVs have the notion of a 'source' that allows one to specify the source of the data. To import data from an external source, the source has to be either 'http' ,'S3', 'GCS' or 'registry'. If your source requires authentication, you can also pass in a `secretRef` to a Kubernetes [Secret](../manifests/example/endpoint-secret.yaml) containing the authentication information.  TLS certificates for https/registry sources may be specified in a [ConfigMap](../manifests/example/cert-configmap.yaml) and referenced by `certConfigMap`.  `secretRef` and `certConfigMap` must be in the same namespace as the DataVolume.

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-import-dv"
spec:
  source:
      http:
         url: "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img" # S3 or GCS
         secretRef: "" # Optional
         certConfigMap: "" # Optional
  storage:
    resources:
      requests:
        storage: "64Mi"
```
[Get example](../manifests/example/import-kubevirt-datavolume.yaml)
[Get GCS example](../manifests/example/import-kubevirt-datavolume-gcs.yaml)
[Get secret example](../manifests/example/endpoint-secret.yaml)
[Get certificate example](../manifests/example/cert-configmap.yaml)

Alternatively, if your certificate is stored in a local file, you can create the `ConfigMap` like this:

```bash
kubectl create configmap import-certs --from-file=ca.pem
```

#### Content-type
You can specify the content type of the source image. The following content-type is valid:
* kubevirt (Virtual disk image, the default if missing)
* archive (Tar archive)
If the content type is kubevirt, the source will be treated as a virtual disk, converted to raw, and sized appropriately. If the content type is archive it will be treated as a tar archive and CDI will attempt to extract the contents of that archive into the Data Volume.
An example of an archive from an http source:

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-import-dv"
spec:
  source:
      http:
         url: "http://server/archive.tar"
         secretRef: "" # Optional
   contentType: "archive"
  storage:
    resources:
      requests:
        storage: "64Mi"
```
#### Extra Headers
You can also specify custom headers directly as a list of strings, with `extraHeaders`:

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-import-dv"
spec:
  source:
      http:
         url: "http://server/archive.tar"
         extraHeaders:
         - "X-First-Header: 12345"
         - "X-Another-Header: abcde"
  storage:
    resources:
      requests:
        storage: "64Mi"
```
If the headers contain information that should not be openly displayed, specify `secretExtraHeaders` as a list of references to secrets:

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-import-dv"
spec:
  source:
      http:
         url: "http://server/archive.tar"
         extraHeaders:
           - "X-Auth-Token: 12345"
           - "X-Another-Header: abcde"
         secretExtraHeaders:
           - "first-secret"
           - "second-secret"
  storage:
    resources:
      requests:
        storage: "64Mi"
```
Each secret should be of type `Opaque`, containing the header within each value under `data` or `stringData`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: first-secret
type: Opaque
stringData:
  secretHeaderOne: "X-Secret-Auth-Token: 6789"
  secretHeaderTwo: "X-Second-Secret-Auth-Token: 5432"
```


### PVC source
You can also use a PVC as an input source for a DV which will cause a clone to happen of the original PVC. You set the 'source' to be PVC, and specify the name and namespace of the PVC you want to have cloned.

The source PVC must not be mounted by any Pod, and CDI will wait as long as it is mounted before proceeding. This check is needed to ensure the snapshot is not corrupted.

Clones can be executed in 2 different ways:
- Host-assisted clones, where Pods exchange the data from one PVC to the other byte by byte. This is the slowest method.
- Smart clones, using the cloning API of the CSI if it is available, see [the documentation](smart-clone.md) for more information.

Smart clones are automatically preferred by CDI, falling back to host-assisted if it cannot use them.

You can also make a clone manually by creating a VolumeSnapshot of a source PVC and creating a DV with a snapshot as a source. This can be done with the source PVC mounted to a Pod, bypassing the check done by CDI, but at the risk of potential corruptions.

Regarding the DV size, CDI can apply some logic to detect the required quantity based on the source PVC, so said amount can be left empty when using the [storage](#storage) API, as shown above:

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-clone-dv"
spec:
  source:
      pvc:
        name: source-pvc
        namespace: example-ns
  storage:
```

However, when using the [pvc](#pvc) API, the user needs to specify the right amount of space to allocate for the new DV, or the clone will not be able to complete.

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-clone-dv"
spec:
  source:
      pvc:
        name: source-pvc
        namespace: example-ns
  pvc:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: "128Mi"  # Size needs to be specified
```
[Get example](../manifests/example/clone-datavolume.yaml)

### VolumeSnapshot source
You can use a volume snapshot as an input source to create a new DV. Set the source to be a snapshot, and specify the name and namespace of the snapshot.

The DV size can be omitted, we apply the same logic as when the source [is a pvc](#pvc-source).

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-snapshot-dv"
spec:
  source:
    snapshot:
      namespace: default
      name: snapshot
  storage:
  # Can be omitted to auto-detect the correct storage size
  # resources:
  #   requests:
  #     storage: 9Gi
```

More details about using snapshots as a source are available [in this document](clone-from-volumesnapshot-source.md).

### Upload Data Volumes
You can upload a virtual disk image directly into a data volume as well, just like with PVCs. The steps to follow are identical as [upload for PVC](upload.md) except that the yaml for a Data Volume is slightly different.
```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: example-upload-dv
spec:
  source:
    upload: {}
  storage:
    resources:
      requests:
        storage: 1Gi
```

### Blank Data Volume
You can create a blank virtual disk image in a Data Volume as well, with the following yaml:
```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: example-blank-dv
spec:
  source:
    blank: {}
  storage:
    resources:
      requests:
        storage: 1Gi
```

### Image IO Data Volume
Image IO sources are sources from oVirt imageio endpoints. In order to use these endpoints you will need an oVirt installation with imageIO enabled. You will then be able to import disk images from oVirt into KubeVirt. The diskId can be obtained from the oVirt webadmin UI or REST api.
```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "test-dv"
spec:
  source:
      imageio:
         url: "http://<ovirt engine url>/ovirt-engine/api"
         secretRef: "endpoint-secret"
         certConfigMap: "tls-certs"
         diskId: "1"
  storage:
    resources:
      requests:
        storage: "500Mi"
```
[Get secret example](../manifests/example/endpoint-secret.yaml)
[Get certificate example](../manifests/example/cert-configmap.yaml)

### VDDK Data Volume
VDDK sources come from VMware vCenter or ESX endpoints. You will need a secret containing administrative credentials for the API provided by the VMware endpoint, as well as a special sidecar image containing the non-redistributable VDDK library folder. Instructions for creating a VDDK image can be found [here](https://docs.openshift.com/container-platform/4.3/cnv/cnv_virtual_machines/cnv_importing_vms/cnv-importing-vmware-vm.html#cnv-creating-vddk-image_cnv-importing-vmware-vm), with the addendum that the ConfigMap should exist in the current CDI namespace and not 'openshift-cnv'. The image URL may also be specified in an optional `initImageURL` field as show below. This field will override the previous ConfigMap.

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "vddk-dv"
spec:
    source:
        vddk:
           backingFile: "[iSCSI_Datastore] vm/vm_1.vmdk" # From 'Hard disk'/'Disk File' in vCenter/ESX VM settings
           url: "https://vcenter.corp.com"
           uuid: "52260566-b032-36cb-55b1-79bf29e30490"
           thumbprint: "20:6C:8A:5D:44:40:B3:79:4B:28:EA:76:13:60:90:6E:49:D9:D9:A3" # SSL fingerprint of vCenter/ESX host
           secretRef: "vddk-credentials"
           initImageURL: "registry:5000/vddk-init:latest"
    storage:
       accessModes:
         - ReadWriteOnce
       resources:
         requests:
           storage: "32Gi"
```
[Get secret example](../manifests/example/endpoint-secret.yaml)
[Get VDDK ConfigMap example](../manifests/example/vddk-configmap.yaml)
[Ways to find thumbprint](https://libguestfs.org/nbdkit-vddk-plugin.1.html#THUMBPRINTS)

#### Extra VDDK Configuration Options

The VDDK library itself looks in a configuration file (such as `/etc/vmware/config`) for extra options to fine tune data transfers. To pass these options through to the VDDK, store the configuration file contents in a ConfigMap with the key `vddk-config-file` and put the name of this ConfigMap in either the `spec.source.vddk.extraArgs` field or a `cdi.kubevirt.io/storage.pod.vddk.extraargs` annotation in the DataVolume specification. The ConfigMap will be mounted to the importer pod as a volume, and the mount directory will have a file named `vddk-config-file` with the contents of the file. This means that the ConfigMap must be placed in the same namespace as the DataVolume, and the ConfigMap should only have one file entry, `vddk-config-file`.

[Example field](../manifests/example/vddk-args-field.yaml)
[Example annotation](../manifests/example/vddk-args-annotation.yaml)
[Example ConfigMap](../manifests/example/vddk-args-configmap.yaml)

## Multi-stage Import
 In a multi-stage import, multiple pods are started in succession to copy different parts of the source to an existing base disk image. Currently only the [ImageIO](#multi-stage-imageio-import) and [VDDK](#multi-stage-vddk-import) data sources support multi-stage imports.

### Multi-stage ImageIO Import
 The ImageIO source allows a warm migration from RHV/oVirt with a snapshot-based multi-stage import. After copying an initial raw disk image as a base, subsequent QCOW snapshots can be applied on top of this base so that only relatively small images need to be downloaded to copy the latest changes from the source. The ImageIO importer downloads each QCOW to scratch space, checks that its backing file matches the expected ID of the previous checkpoint, then rebases and commits the image to the previously-downloaded image in the PV.

 To create a multi-stage ImageIO import, find the ID of the disk to import along with the IDs of each snapshot that needs to be transferred. Add these IDs as in the following example CRD, and set the finalCheckpoint flag if no further snapshots will be downloaded. The DataVolume will go through the usual import sequence and move to the "Paused" state after each checkpoint. If the finalCheckpoint flag was set to true, the DataVolume will move to "Succeeded" after importing the last checkpoint specified in this list.

 ```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "imageio"
 spec:
  source:
    imageio:
      url: "https://rhv.example.local/ovirt-engine/api"
      secretRef: "endpoint-secret"
      certConfigMap: "tls-certs"
      diskId: "3406e724-7d02-4225-a620-3e6ef646c68c"
  finalCheckpoint: false
  checkpoints:
    - previous: ""
      current: "1c44c27e-d2d8-49c4-841a-cc26c4b1e406"
    - previous: "1c44c27e-d2d8-49c4-841a-cc26c4b1e406"
      current: "c55bb7bb-20f2-46b5-a7f3-11fd6010b7d0"
  storage:
    resources:
      requests:
        storage: "32Gi"
 ```

### Multi-stage VDDK Import
 The VDDK source uses a multi-stage import to perform warm migration: after copying an initial disk image, it queries the VMware host for the blocks that changed in between two snapshots. Each delta is applied to the disk image, and only the final delta copy needs the source VM to be powered off, minimizing downtime.

To create a multi-stage VDDK import, first [enable changed block tracking](https://kb.vmware.com/s/article/1031873) on the source VM. Take an initial snapshot of the VM (snapshot-1), and take another snapshot (snapshot-2) after the VM has run long enough to save more data to disk. Create a DataVolume spec similar to the example below, specifying a list of checkpoints and a finalCheckpoint boolean to indicate if there are no further snapshots to copy. The first importer pod to appear will copy the full disk contents of snapshot-1 to the disk image provided by the PVC, and the second importer pod will quickly copy only the blocks that changed between snapshot-1 and snapshot-2. If finalCheckpoint is set to false, the resulting DataVolume will wait in a "Paused" state until further checkpoints are provided. The DataVolume will only move to "Succeeded" when finalCheckpoint is true and the last checkpoint in the list has been copied. It is not necessary to provide all the checkpoints up-front, because updates are allowed to be applied to these fields (finalCheckpoint and checkpoints).

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "vddk-multistage-dv"
spec:
    source:
        vddk:
           backingFile: "[iSCSI_Datastore] vm/vm_1.vmdk" # From latest 'Hard disk'/'Disk File' in vCenter/ESX VM settings
           url: "https://vcenter.corp.com"
           uuid: "52260566-b032-36cb-55b1-79bf29e30490"
           thumbprint: "20:6C:8A:5D:44:40:B3:79:4B:28:EA:76:13:60:90:6E:49:D9:D9:A3" # SSL fingerprint of vCenter/ESX host
           secretRef: "vddk-credentials"
    finalCheckpoint: true
    checkpoints:
      - current: "snapshot-1"
        previous: ""
      - current: "snapshot-2"
        previous: "snapshot-1"
    storage:
       accessModes:
         - ReadWriteOnce
       resources:
         requests:
           storage: "32Gi"
```

#### Change IDs
For multi-stage VDDK imports, the `previous` field in the list of checkpoints can be specified as a snapshot ID or a change ID. A change ID persists after its associated snapshot has been deleted, so it can be used to implement a warm migration workflow that does not leave chains of snapshots that need to be cleaned (see [this VMware KB article](https://kb.vmware.com/s/article/76082)).

Example: initiate a multi-stage VDDK import, specifying the initial snapshot as the first checkpoint.

```yaml
finalCheckpoint: false
checkpoints:
  - current: "snapshot-1"
    previous: ""
```

After the snapshot has been copied and the Data Volume is `Paused`, save the snapshot's change ID, delete the snapshot, and let the VM run for some time so that the data on the disk changes somewhat. Take a new snapshot, and append the new checkpoint to the list using the new snapshot ID as `current` and the deleted snapshot's change ID as `previous`:

```yaml
finalCheckpoint: false
checkpoints:
  - current: "snapshot-1"
    previous: ""
  - current: "snapshot-2"
    previous: "53 d0 ac 95 4f 09 f7 93-b1 21 e2 39 97 8a fa 63/4"
```

This process can be repeated until the VM can be shut down for a final snapshot copy with `finalCheckpoint` set to `true`.

## Conditions
The DataVolume status object has conditions. There are 3 conditions available for DataVolumes
* Ready
* Bound
* Running

The running and ready conditions are mutually exclusive, if running is true, then ready cannot be true and vice versa. Each condition has the following fields:
* Type (Ready/Bound/Running).
* Status (True/False).
* LastTransitionTime - the timestamp when the last transition happened.
* LastHeartbeatTime - the timestamp the last time anything on the condition was updated.
* Reason - the reason the status transitioned to a new value, this is a camel cased single word, similar to an EventReason in events.
* Message - a detailed messages expanding on the reason of the transition. For instance if Running went from True to False, the reason will be the container exit reason, and the message will be the container exit message, which explains why the container exited.

## Annotations
Specific [DV annotations](datavolume-annotations.md) are passed to the transfer pods to control their behavior.
Other [annotations](debug.md) help debugging and testing by retaining the transfer pods after completion.

## Priority Class
You can specify priority class name on the Data Volume Object. The corresponding pod created for the data volume will be assigned the priority class on the data volume. Following is an example of specifying the priority class on Data Volume 
```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: "example-priority-class-dv"
spec:
  priorityClassName: kubevirt
  source:
    ...
  storage:
    ...
```

## Kubevirt integration
[Kubevirt](https://github.com/kubevirt/kubevirt) is an extension to Kubernetes that allows one to run Virtual Machines(VM) on the same infra structure as the containers managed by Kubernetes. CDI provides a mechanism to get a disk image into a PVC in order for Kubevirt to consume it. The following steps have to be taken in order for Kubevirt to consume a CDI provided disk image.
1. Create a PVC with an annotation to for instance import from an external URL.
2. An importer pod is started, that attempts to get the image from the external source.
3. Create a VM definition that references the PVC we just created.
4. Wait for the importer pod to finish (status can be checked by the status annotation on the PVC).
5. Start the VMs using the imported disk.
There is no mechanism to stop 5 from happening before the import is complete, so one can attempt to start the VM before the disk has been completely imported, with obvious bad results.

Now let's do the same process but using DVs.
1. Create a VM definition that references a DV template, which includes the external URL that contains the disk image.
2. A DV is created from the template that in turn creates an underlying PVC with the correct annotation.
3. The importer pod is created like before.
4. Until the DV status is Success, the virt launcher controller will not schedule the VM to be launched if the user tries to start the VM.
We now have a fully controlled mechanism where we can define a VM using a DV with a disk image from an external source, that cannot be scheduled to run until the import has been completed.

### Example VM using DV
```yaml
apiVersion: kubevirt.io/v1alpha3
kind: VirtualMachine
metadata:
  creationTimestamp: null
  labels:
    kubevirt.io/vm: vm-fedora-datavolume
  name: vm-fedora-datavolume
spec:
  dataVolumeTemplates:
  - metadata:
      creationTimestamp: null
      name: fedora-dv
    spec:
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 100M
        storageClassName: hdd
      source:
        http:
          url: https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img
  running: false
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-datavolume
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: datavolumevolume
        machine:
          type: ""
        resources:
          requests:
            memory: 64M
      terminationGracePeriodSeconds: 0
      volumes:
      - dataVolume:
          name: fedora-dv
        name: datavolumevolume
```
[Get example](../manifests/example/vm-dv.yaml)

This example combines all the different pieces into a single yaml.
* Creation of a VM definition (example-vm).
* Creation of a DV with a source of http which points to an external URL (example-dv).
* Creation of a matching PVC with the same name as the DV, which will contain the result of the import (example-dv).
* Creation of an importer pod that does the actual import work.
