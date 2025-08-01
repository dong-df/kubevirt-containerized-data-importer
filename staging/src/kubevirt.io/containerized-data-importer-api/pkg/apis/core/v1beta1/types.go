/*
Copyright 2018 The CDI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/api"
)

// DataVolume is an abstraction on top of PersistentVolumeClaims to allow easy population of those PersistentVolumeClaims with relation to VirtualMachines
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=dv;dvs,categories=all
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The phase the data volume is in"
// +kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".status.progress",description="Transfer progress in percentage if known, N/A otherwise"
// +kubebuilder:printcolumn:name="Restarts",type="integer",JSONPath=".status.restartCount",description="The number of times the transfer has been restarted."
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DataVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DataVolumeSpec `json:"spec"`
	// +optional
	Status DataVolumeStatus `json:"status,omitempty"`
}

// DataVolumeSpec defines the DataVolume type specification
type DataVolumeSpec struct {
	//Source is the src of the data for the requested DataVolume
	// +optional
	Source *DataVolumeSource `json:"source,omitempty"`
	//SourceRef is an indirect reference to the source of data for the requested DataVolume
	// +optional
	SourceRef *DataVolumeSourceRef `json:"sourceRef,omitempty"`
	//PVC is the PVC specification
	PVC *corev1.PersistentVolumeClaimSpec `json:"pvc,omitempty"`
	// Storage is the requested storage specification
	Storage *StorageSpec `json:"storage,omitempty"`
	//PriorityClassName for Importer, Cloner and Uploader pod
	PriorityClassName string `json:"priorityClassName,omitempty"`
	//DataVolumeContentType options: "kubevirt", "archive"
	// +kubebuilder:validation:Enum="kubevirt";"archive"
	ContentType DataVolumeContentType `json:"contentType,omitempty"`
	// Checkpoints is a list of DataVolumeCheckpoints, representing stages in a multistage import.
	Checkpoints []DataVolumeCheckpoint `json:"checkpoints,omitempty"`
	// FinalCheckpoint indicates whether the current DataVolumeCheckpoint is the final checkpoint.
	FinalCheckpoint bool `json:"finalCheckpoint,omitempty"`
	// Preallocation controls whether storage for DataVolumes should be allocated in advance.
	Preallocation *bool `json:"preallocation,omitempty"`
}

// StorageSpec defines the Storage type specification
type StorageSpec struct {
	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	// A label query over volumes to consider for binding.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// Resources represents the minimum resources the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	// +optional
	Resources corev1.VolumeResourceRequirements `json:"resources,omitempty"`
	// VolumeName is the binding reference to the PersistentVolume backing this claim.
	// +optional
	VolumeName string `json:"volumeName,omitempty"`
	// Name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
	// volumeMode defines what type of volume is required by the claim.
	// Value of Filesystem is implied when not included in claim spec.
	// +optional
	VolumeMode *corev1.PersistentVolumeMode `json:"volumeMode,omitempty"`
	// This field can be used to specify either: * An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An existing PVC (PersistentVolumeClaim) * An existing custom resource that implements data population (Alpha) In order to use custom resource types that implement data population, the AnyVolumeDataSource feature gate must be enabled. If the provisioner or an external controller can support the specified data source, it will create a new volume based on the contents of the specified data source.
	// If the AnyVolumeDataSource feature gate is enabled, this field will always have the same contents as the DataSourceRef field.
	// +optional
	DataSource *corev1.TypedLocalObjectReference `json:"dataSource,omitempty"`
	// Specifies the object from which to populate the volume with data, if a non-empty volume is desired. This may be any local object from a non-empty API group (non core object) or a PersistentVolumeClaim object. When this field is specified, volume binding will only succeed if the type of the specified object matches some installed volume populator or dynamic provisioner.
	// This field will replace the functionality of the DataSource field and as such if both fields are non-empty, they must have the same value. For backwards compatibility, both fields (DataSource and DataSourceRef) will be set to the same value automatically if one of them is empty and the other is non-empty.
	// There are two important differences between DataSource and DataSourceRef:
	// * While DataSource only allows two specific types of objects, DataSourceRef allows any non-core object, as well as PersistentVolumeClaim objects.
	// * While DataSource ignores disallowed values (dropping them), DataSourceRef preserves all values, and generates an error if a disallowed value is specified.
	// (Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
	// +optional
	DataSourceRef *corev1.TypedObjectReference `json:"dataSourceRef,omitempty"`
}

// PersistentVolumeFromStorageProfile means the volume mode will be auto selected by CDI according to a matching StorageProfile
const PersistentVolumeFromStorageProfile corev1.PersistentVolumeMode = "FromStorageProfile"

// DataVolumeCheckpoint defines a stage in a warm migration.
type DataVolumeCheckpoint struct {
	// Previous is the identifier of the snapshot from the previous checkpoint.
	Previous string `json:"previous"`
	// Current is the identifier of the snapshot created for this checkpoint.
	Current string `json:"current"`
}

// DataVolumeContentType represents the types of the imported data
type DataVolumeContentType string

const (
	// DataVolumeKubeVirt is the content-type of the imported file, defaults to kubevirt
	DataVolumeKubeVirt DataVolumeContentType = "kubevirt"
	// DataVolumeArchive is the content-type to specify if there is a need to extract the imported archive
	DataVolumeArchive DataVolumeContentType = "archive"
)

// DataVolumeSource represents the source for our Data Volume, this can be HTTP, Imageio, S3, GCS, Registry or an existing PVC
type DataVolumeSource struct {
	HTTP     *DataVolumeSourceHTTP     `json:"http,omitempty"`
	S3       *DataVolumeSourceS3       `json:"s3,omitempty"`
	GCS      *DataVolumeSourceGCS      `json:"gcs,omitempty"`
	Registry *DataVolumeSourceRegistry `json:"registry,omitempty"`
	PVC      *DataVolumeSourcePVC      `json:"pvc,omitempty"`
	Upload   *DataVolumeSourceUpload   `json:"upload,omitempty"`
	Blank    *DataVolumeBlankImage     `json:"blank,omitempty"`
	Imageio  *DataVolumeSourceImageIO  `json:"imageio,omitempty"`
	VDDK     *DataVolumeSourceVDDK     `json:"vddk,omitempty"`
	Snapshot *DataVolumeSourceSnapshot `json:"snapshot,omitempty"`
}

// DataVolumeSourcePVC provides the parameters to create a Data Volume from an existing PVC
type DataVolumeSourcePVC struct {
	// The namespace of the source PVC
	Namespace string `json:"namespace"`
	// The name of the source PVC
	Name string `json:"name"`
}

// DataVolumeSourceSnapshot provides the parameters to create a Data Volume from an existing VolumeSnapshot
type DataVolumeSourceSnapshot struct {
	// The namespace of the source VolumeSnapshot
	Namespace string `json:"namespace"`
	// The name of the source VolumeSnapshot
	Name string `json:"name"`
}

// DataSourceRefSourceDataSource serves as a reference to another DataSource
// Can be resolved into a DataVolumeSourcePVC or a DataVolumeSourceSnapshot
// The maximum depth of a reference chain may not exceed 1.
type DataSourceRefSourceDataSource struct {
	// The namespace of the source DataSource
	Namespace string `json:"namespace"`
	// The name of the source DataSource
	Name string `json:"name"`

}

// DataVolumeBlankImage provides the parameters to create a new raw blank image for the PVC
type DataVolumeBlankImage struct{}

// DataVolumeSourceUpload provides the parameters to create a Data Volume by uploading the source
type DataVolumeSourceUpload struct {
}

// DataVolumeSourceS3 provides the parameters to create a Data Volume from an S3 source
type DataVolumeSourceS3 struct {
	//URL is the url of the S3 source
	URL string `json:"url"`
	//SecretRef provides the secret reference needed to access the S3 source
	SecretRef string `json:"secretRef,omitempty"`
	// CertConfigMap is a configmap reference, containing a Certificate Authority(CA) public key, and a base64 encoded pem certificate
	// +optional
	CertConfigMap string `json:"certConfigMap,omitempty"`
}

// DataVolumeSourceGCS provides the parameters to create a Data Volume from an GCS source
type DataVolumeSourceGCS struct {
	//URL is the url of the GCS source
	URL string `json:"url"`
	//SecretRef provides the secret reference needed to access the GCS source
	SecretRef string `json:"secretRef,omitempty"`
}

// DataVolumeSourceRegistry provides the parameters to create a Data Volume from an registry source
type DataVolumeSourceRegistry struct {
	//URL is the url of the registry source (starting with the scheme: docker, oci-archive)
	// +optional
	URL *string `json:"url,omitempty"`
	//ImageStream is the name of image stream for import
	// +optional
	ImageStream *string `json:"imageStream,omitempty"`
	//PullMethod can be either "pod" (default import), or "node" (node docker cache based import)
	// +optional
	PullMethod *RegistryPullMethod `json:"pullMethod,omitempty"`
	//SecretRef provides the secret reference needed to access the Registry source
	// +optional
	SecretRef *string `json:"secretRef,omitempty"`
	//CertConfigMap provides a reference to the Registry certs
	// +optional
	CertConfigMap *string `json:"certConfigMap,omitempty"`
	//Platform describes the minimum runtime requirements of the image
	// +optional
	Platform *PlatformOptions `json:"platform,omitempty"`
}

type PlatformOptions struct {
	//Architecture specifies the image target CPU architecture
	// +optional
	Architecture string `json:"architecture,omitempty"`
}

const (
	// RegistrySchemeDocker is docker scheme prefix
	RegistrySchemeDocker = "docker"
	// RegistrySchemeOci is oci-archive scheme prefix
	RegistrySchemeOci = "oci-archive"
)

// RegistryPullMethod represents the registry import pull method
type RegistryPullMethod string

const (
	// RegistryPullPod is the standard import
	RegistryPullPod RegistryPullMethod = "pod"
	// RegistryPullNode is the node docker cache based import
	RegistryPullNode RegistryPullMethod = "node"
)

// DataVolumeSourceHTTP can be either an http or https endpoint, with an optional basic auth user name and password, and an optional configmap containing additional CAs
type DataVolumeSourceHTTP struct {
	// URL is the URL of the http(s) endpoint
	URL string `json:"url"`
	// SecretRef A Secret reference, the secret should contain accessKeyId (user name) base64 encoded, and secretKey (password) also base64 encoded
	// +optional
	SecretRef string `json:"secretRef,omitempty"`
	// CertConfigMap is a configmap reference, containing a Certificate Authority(CA) public key, and a base64 encoded pem certificate
	// +optional
	CertConfigMap string `json:"certConfigMap,omitempty"`
	// ExtraHeaders is a list of strings containing extra headers to include with HTTP transfer requests
	// +optional
	ExtraHeaders []string `json:"extraHeaders,omitempty"`
	// SecretExtraHeaders is a list of Secret references, each containing an extra HTTP header that may include sensitive information
	// +optional
	SecretExtraHeaders []string `json:"secretExtraHeaders,omitempty"`
}

// DataVolumeSourceImageIO provides the parameters to create a Data Volume from an imageio source
type DataVolumeSourceImageIO struct {
	//URL is the URL of the ovirt-engine
	URL string `json:"url"`
	// DiskID provides id of a disk to be imported
	DiskID string `json:"diskId"`
	//SecretRef provides the secret reference needed to access the ovirt-engine
	SecretRef string `json:"secretRef,omitempty"`
	//CertConfigMap provides a reference to the CA cert
	CertConfigMap string `json:"certConfigMap,omitempty"`
}

// DataVolumeSourceVDDK provides the parameters to create a Data Volume from a Vmware source
type DataVolumeSourceVDDK struct {
	// URL is the URL of the vCenter or ESXi host with the VM to migrate
	URL string `json:"url,omitempty"`
	// UUID is the UUID of the virtual machine that the backing file is attached to in vCenter/ESXi
	UUID string `json:"uuid,omitempty"`
	// BackingFile is the path to the virtual hard disk to migrate from vCenter/ESXi
	BackingFile string `json:"backingFile,omitempty"`
	// Thumbprint is the certificate thumbprint of the vCenter or ESXi host
	Thumbprint string `json:"thumbprint,omitempty"`
	// SecretRef provides a reference to a secret containing the username and password needed to access the vCenter or ESXi host
	SecretRef string `json:"secretRef,omitempty"`
	// InitImageURL is an optional URL to an image containing an extracted VDDK library, overrides v2v-vmware config map
	InitImageURL string `json:"initImageURL,omitempty"`
	// ExtraArgs is a reference to a ConfigMap containing extra arguments to pass directly to the VDDK library
	ExtraArgs string `json:"extraArgs,omitempty"`
}

// DataVolumeSourceRef defines an indirect reference to the source of data for the DataVolume
type DataVolumeSourceRef struct {
	// The kind of the source reference, currently only "DataSource" is supported
	Kind string `json:"kind"`
	// The namespace of the source reference, defaults to the DataVolume namespace
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// The name of the source reference
	Name string `json:"name"`
}

const (
	// DataVolumeDataSource is DataSource source reference for DataVolume
	DataVolumeDataSource = "DataSource"
)

// DataVolumeStatus contains the current status of the DataVolume
type DataVolumeStatus struct {
	// ClaimName is the name of the underlying PVC used by the DataVolume.
	ClaimName string `json:"claimName,omitempty"`
	//Phase is the current phase of the data volume
	Phase    DataVolumePhase    `json:"phase,omitempty"`
	Progress DataVolumeProgress `json:"progress,omitempty"`
	// RestartCount is the number of times the pod populating the DataVolume has restarted
	RestartCount int32                 `json:"restartCount,omitempty"`
	Conditions   []DataVolumeCondition `json:"conditions,omitempty" optional:"true"`
}

// DataVolumeList provides the needed parameters to do request a list of Data Volumes from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of DataVolumes
	Items []DataVolume `json:"items"`
}

// DataVolumeCondition represents the state of a data volume condition.
type DataVolumeCondition struct {
	Type               DataVolumeConditionType `json:"type" description:"type of condition ie. Ready|Bound|Running."`
	Status             corev1.ConditionStatus  `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time             `json:"lastTransitionTime,omitempty"`
	LastHeartbeatTime  metav1.Time             `json:"lastHeartbeatTime,omitempty"`
	Reason             string                  `json:"reason,omitempty" description:"reason for the condition's last transition"`
	Message            string                  `json:"message,omitempty" description:"human-readable message indicating details about last transition"`
}

// DataVolumePhase is the current phase of the DataVolume
type DataVolumePhase string

// DataVolumeProgress is the current progress of the DataVolume transfer operation. Value between 0 and 100 inclusive, N/A if not available
type DataVolumeProgress string

// DataVolumeConditionType is the string representation of known condition types
type DataVolumeConditionType string

const (
	// PhaseUnset represents a data volume with no current phase
	PhaseUnset DataVolumePhase = ""

	// Pending represents a data volume with a current phase of Pending
	Pending DataVolumePhase = "Pending"
	// PVCBound represents a data volume with a current phase of PVCBound
	PVCBound DataVolumePhase = "PVCBound"

	// ImportScheduled represents a data volume with a current phase of ImportScheduled
	ImportScheduled DataVolumePhase = "ImportScheduled"

	// ImportInProgress represents a data volume with a current phase of ImportInProgress
	ImportInProgress DataVolumePhase = "ImportInProgress"

	// CloneScheduled represents a data volume with a current phase of CloneScheduled
	CloneScheduled DataVolumePhase = "CloneScheduled"

	// CloneInProgress represents a data volume with a current phase of CloneInProgress
	CloneInProgress DataVolumePhase = "CloneInProgress"

	// SnapshotForSmartCloneInProgress represents a data volume with a current phase of SnapshotForSmartCloneInProgress
	SnapshotForSmartCloneInProgress DataVolumePhase = "SnapshotForSmartCloneInProgress"

	// CloneFromSnapshotSourceInProgress represents a data volume with a current phase of CloneFromSnapshotSourceInProgress
	CloneFromSnapshotSourceInProgress DataVolumePhase = "CloneFromSnapshotSourceInProgress"

	// SmartClonePVCInProgress represents a data volume with a current phase of SmartClonePVCInProgress
	SmartClonePVCInProgress DataVolumePhase = "SmartClonePVCInProgress"

	// CSICloneInProgress represents a data volume with a current phase of CSICloneInProgress
	CSICloneInProgress DataVolumePhase = "CSICloneInProgress"

	// ExpansionInProgress is the state when a PVC is expanded
	ExpansionInProgress DataVolumePhase = "ExpansionInProgress"

	// NamespaceTransferInProgress is the state when a PVC is transferred
	NamespaceTransferInProgress DataVolumePhase = "NamespaceTransferInProgress"

	// UploadScheduled represents a data volume with a current phase of UploadScheduled
	UploadScheduled DataVolumePhase = "UploadScheduled"

	// UploadReady represents a data volume with a current phase of UploadReady
	UploadReady DataVolumePhase = "UploadReady"

	// WaitForFirstConsumer represents a data volume with a current phase of WaitForFirstConsumer
	WaitForFirstConsumer DataVolumePhase = "WaitForFirstConsumer"
	// PendingPopulation represents a data volume which should be populated by
	// the CDI populators but haven't created the pvc' yet
	PendingPopulation DataVolumePhase = "PendingPopulation"

	// Succeeded represents a DataVolumePhase of Succeeded
	Succeeded DataVolumePhase = "Succeeded"
	// Failed represents a DataVolumePhase of Failed
	Failed DataVolumePhase = "Failed"
	// Unknown represents a DataVolumePhase of Unknown
	Unknown DataVolumePhase = "Unknown"
	// Paused represents a DataVolumePhase of Paused
	Paused DataVolumePhase = "Paused"

	// PrepClaimInProgress represents a data volume with a current phase of PrepClaimInProgress
	PrepClaimInProgress DataVolumePhase = "PrepClaimInProgress"
	// RebindInProgress represents a data volume with a current phase of RebindInProgress
	RebindInProgress DataVolumePhase = "RebindInProgress"

	// DataVolumeReady is the condition that indicates if the data volume is ready to be consumed.
	DataVolumeReady DataVolumeConditionType = "Ready"
	// DataVolumeBound is the condition that indicates if the underlying PVC is bound or not.
	DataVolumeBound DataVolumeConditionType = "Bound"
	// DataVolumeRunning is the condition that indicates if the import/upload/clone container is running.
	DataVolumeRunning DataVolumeConditionType = "Running"
)

// DataVolumeCloneSourceSubresource is the subresource checked for permission to clone
const DataVolumeCloneSourceSubresource = "source"

// this has to be here otherwise informer-gen doesn't recognize it
// see https://github.com/kubernetes/code-generator/issues/59
// +genclient:nonNamespaced

// StorageProfile provides a CDI specific recommendation for storage parameters
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster
type StorageProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageProfileSpec   `json:"spec"`
	Status StorageProfileStatus `json:"status,omitempty"`
}

// StorageProfileSpec defines specification for StorageProfile
type StorageProfileSpec struct {
	// CloneStrategy defines the preferred method for performing a CDI clone
	CloneStrategy *CDICloneStrategy `json:"cloneStrategy,omitempty"`
	// ClaimPropertySets is a provided set of properties applicable to PVC
	// +kubebuilder:validation:MaxItems=8
	ClaimPropertySets []ClaimPropertySet `json:"claimPropertySets,omitempty"`
	// DataImportCronSourceFormat defines the format of the DataImportCron-created disk image sources
	DataImportCronSourceFormat *DataImportCronSourceFormat `json:"dataImportCronSourceFormat,omitempty"`
	// SnapshotClass is optional specific VolumeSnapshotClass for CloneStrategySnapshot. If not set, a VolumeSnapshotClass is chosen according to the provisioner.
	SnapshotClass *string `json:"snapshotClass,omitempty"`
}

// StorageProfileStatus provides the most recently observed status of the StorageProfile
type StorageProfileStatus struct {
	// The StorageClass name for which capabilities are defined
	StorageClass *string `json:"storageClass,omitempty"`
	// The Storage class provisioner plugin name
	Provisioner *string `json:"provisioner,omitempty"`
	// CloneStrategy defines the preferred method for performing a CDI clone
	CloneStrategy *CDICloneStrategy `json:"cloneStrategy,omitempty"`
	// ClaimPropertySets computed from the spec and detected in the system
	// +kubebuilder:validation:MaxItems=8
	ClaimPropertySets []ClaimPropertySet `json:"claimPropertySets,omitempty"`
	// DataImportCronSourceFormat defines the format of the DataImportCron-created disk image sources
	DataImportCronSourceFormat *DataImportCronSourceFormat `json:"dataImportCronSourceFormat,omitempty"`
	// SnapshotClass is optional specific VolumeSnapshotClass for CloneStrategySnapshot. If not set, a VolumeSnapshotClass is chosen according to the provisioner.
	SnapshotClass *string `json:"snapshotClass,omitempty"`
}

// ClaimPropertySet is a set of properties applicable to PVC
type ClaimPropertySet struct {
	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +kubebuilder:validation:MaxItems=4
	// +kubebuilder:validation:XValidation:rule="self.all(am, am in ['ReadWriteOnce', 'ReadOnlyMany', 'ReadWriteMany', 'ReadWriteOncePod'])", message="Illegal AccessMode"
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes"`
	// VolumeMode defines what type of volume is required by the claim.
	// Value of Filesystem is implied when not included in claim spec.
	// +kubebuilder:validation:Enum="Block";"Filesystem"
	VolumeMode *corev1.PersistentVolumeMode `json:"volumeMode"`
}

// StorageProfileList provides the needed parameters to request a list of StorageProfile from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StorageProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of StorageProfile
	Items []StorageProfile `json:"items"`
}

// DataSource references an import/clone source for a DataVolume
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=das,categories=all
type DataSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSourceSpec   `json:"spec"`
	Status DataSourceStatus `json:"status,omitempty"`
}

// DataSourceSpec defines specification for DataSource
type DataSourceSpec struct {
	// Source is the source of the data referenced by the DataSource
	Source DataSourceSource `json:"source"`
}

// DataSourceSource represents the source for our DataSource
type DataSourceSource struct {
	// +optional
	PVC *DataVolumeSourcePVC `json:"pvc,omitempty"`
	// +optional
	Snapshot *DataVolumeSourceSnapshot `json:"snapshot,omitempty"`
	// +optional
	DataSource *DataSourceRefSourceDataSource `json:"dataSource,omitempty"` 
}

// DataSourceStatus provides the most recently observed status of the DataSource
type DataSourceStatus struct {
	// Source is the current source of the data referenced by the DataSource
	Source     DataSourceSource      `json:"source,omitempty"`
	Conditions []DataSourceCondition `json:"conditions,omitempty" optional:"true"`
}

// DataSourceCondition represents the state of a data source condition
type DataSourceCondition struct {
	Type           DataSourceConditionType `json:"type" description:"type of condition ie. Ready"`
	ConditionState `json:",inline"`
}

// DataSourceConditionType is the string representation of known condition types
type DataSourceConditionType string

const (
	// DataSourceReady is the condition that indicates if the data source is ready to be consumed
	DataSourceReady DataSourceConditionType = "Ready"
)

// ConditionState represents the state of a condition
type ConditionState struct {
	Status             corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	LastHeartbeatTime  metav1.Time            `json:"lastHeartbeatTime,omitempty"`
	Reason             string                 `json:"reason,omitempty" description:"reason for the condition's last transition"`
	Message            string                 `json:"message,omitempty" description:"human-readable message indicating details about last transition"`
}

// DataSourceList provides the needed parameters to do request a list of Data Sources from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of DataSources
	Items []DataSource `json:"items"`
}

// DataImportCron defines a cron job for recurring polling/importing disk images as PVCs into a golden image namespace
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=dic;dics,categories=all
// +kubebuilder:printcolumn:name="Format",type="string",JSONPath=".status.sourceFormat",description="The format in which created sources are saved"
type DataImportCron struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataImportCronSpec   `json:"spec"`
	Status DataImportCronStatus `json:"status,omitempty"`
}

// DataImportCronSpec defines specification for DataImportCron
type DataImportCronSpec struct {
	// Template specifies template for the DVs to be created
	Template DataVolume `json:"template"`
	// Schedule specifies in cron format when and how often to look for new imports
	Schedule string `json:"schedule"`
	// GarbageCollect specifies whether old PVCs should be cleaned up after a new PVC is imported.
	// Options are currently "Outdated" and "Never", defaults to "Outdated".
	// +optional
	GarbageCollect *DataImportCronGarbageCollect `json:"garbageCollect,omitempty"`
	// Number of import PVCs to keep when garbage collecting. Default is 3.
	// +optional
	ImportsToKeep *int32 `json:"importsToKeep,omitempty"`
	// ManagedDataSource specifies the name of the corresponding DataSource this cron will manage.
	// DataSource has to be in the same namespace.
	ManagedDataSource string `json:"managedDataSource"`
	// RetentionPolicy specifies whether the created DataVolumes and DataSources are retained when their DataImportCron is deleted. Default is RatainAll.
	// +optional
	RetentionPolicy *DataImportCronRetentionPolicy `json:"retentionPolicy,omitempty"`
}

// DataImportCronGarbageCollect represents the DataImportCron garbage collection mode
type DataImportCronGarbageCollect string

const (
	// DataImportCronGarbageCollectNever specifies that garbage collection is disabled
	DataImportCronGarbageCollectNever DataImportCronGarbageCollect = "Never"
	// DataImportCronGarbageCollectOutdated specifies that old PVCs should be cleaned up after a new PVC is imported
	DataImportCronGarbageCollectOutdated DataImportCronGarbageCollect = "Outdated"
)

// DataImportCronRetentionPolicy represents the DataImportCron retention policy
type DataImportCronRetentionPolicy string

const (
	// DataImportCronRetainNone specifies that the created DataVolumes and DataSources are deleted when their DataImportCron is deleted
	DataImportCronRetainNone DataImportCronRetentionPolicy = "None"
	// DataImportCronRetainAll specifies that the created DataVolumes and DataSources are retained when their DataImportCron is deleted
	DataImportCronRetainAll DataImportCronRetentionPolicy = "All"
)

// DataImportCronStatus provides the most recently observed status of the DataImportCron
type DataImportCronStatus struct {
	// CurrentImports are the imports in progress. Currently only a single import is supported.
	CurrentImports []ImportStatus `json:"currentImports,omitempty"`
	// LastImportedPVC is the last imported PVC
	LastImportedPVC *DataVolumeSourcePVC `json:"lastImportedPVC,omitempty"`
	// LastExecutionTimestamp is the time of the last polling
	LastExecutionTimestamp *metav1.Time `json:"lastExecutionTimestamp,omitempty"`
	// LastImportTimestamp is the time of the last import
	LastImportTimestamp *metav1.Time `json:"lastImportTimestamp,omitempty"`
	// SourceFormat defines the format of the DataImportCron-created disk image sources
	SourceFormat *DataImportCronSourceFormat `json:"sourceFormat,omitempty"`
	Conditions   []DataImportCronCondition   `json:"conditions,omitempty" optional:"true"`
}

// ImportStatus of a currently in progress import
type ImportStatus struct {
	// DataVolumeName is the currently in progress import DataVolume
	DataVolumeName string `json:"DataVolumeName"`
	// Digest of the currently imported image
	Digest string `json:"Digest"`
}

// DataImportCronCondition represents the state of a data import cron condition
type DataImportCronCondition struct {
	Type           DataImportCronConditionType `json:"type" description:"type of condition ie. Progressing, UpToDate"`
	ConditionState `json:",inline"`
}

// DataImportCronConditionType is the string representation of known condition types
type DataImportCronConditionType string

const (
	// DataImportCronProgressing is the condition that indicates import is progressing
	DataImportCronProgressing DataImportCronConditionType = "Progressing"

	// DataImportCronUpToDate is the condition that indicates latest import is up to date
	DataImportCronUpToDate DataImportCronConditionType = "UpToDate"
)

// DataImportCronList provides the needed parameters to do request a list of DataImportCrons from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataImportCronList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of DataImportCrons
	Items []DataImportCron `json:"items"`
}

// VolumeImportSource works as a specification to populate PersistentVolumeClaims with data
// imported from an HTTP/S3/Registry/Blank/ImageIO/VDDK source
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
type VolumeImportSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumeImportSourceSpec `json:"spec"`
	// +optional
	Status VolumeImportSourceStatus `json:"status"`
}

// VolumeImportSourceSpec defines the Spec field for VolumeImportSource
type VolumeImportSourceSpec struct {
	//Source is the src of the data to be imported in the target PVC
	Source *ImportSourceType `json:"source,omitempty"`
	// Preallocation controls whether storage for the target PVC should be allocated in advance.
	Preallocation *bool `json:"preallocation,omitempty"`
	// ContentType represents the type of the imported data (Kubevirt or archive)
	ContentType DataVolumeContentType `json:"contentType,omitempty"`
	// TargetClaim the name of the specific claim to be populated with a multistage import.
	TargetClaim *string `json:"targetClaim,omitempty"`
	// Checkpoints is a list of DataVolumeCheckpoints, representing stages in a multistage import.
	Checkpoints []DataVolumeCheckpoint `json:"checkpoints,omitempty"`
	// FinalCheckpoint indicates whether the current DataVolumeCheckpoint is the final checkpoint.
	FinalCheckpoint *bool `json:"finalCheckpoint,omitempty"`
}

// ImportSourceType contains each one of the source types allowed in a VolumeImportSource
type ImportSourceType struct {
	HTTP     *DataVolumeSourceHTTP     `json:"http,omitempty"`
	S3       *DataVolumeSourceS3       `json:"s3,omitempty"`
	Registry *DataVolumeSourceRegistry `json:"registry,omitempty"`
	GCS      *DataVolumeSourceGCS      `json:"gcs,omitempty"`
	Blank    *DataVolumeBlankImage     `json:"blank,omitempty"`
	Imageio  *DataVolumeSourceImageIO  `json:"imageio,omitempty"`
	VDDK     *DataVolumeSourceVDDK     `json:"vddk,omitempty"`
}

// VolumeImportSourceStatus provides the most recently observed status of the VolumeImportSource
type VolumeImportSourceStatus struct {
}

// VolumeImportSourceList provides the needed parameters to do request a list of Import Sources from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VolumeImportSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of DataSources
	Items []VolumeImportSource `json:"items"`
}

// VolumeUploadSource is a specification to populate PersistentVolumeClaims with upload data
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
type VolumeUploadSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumeUploadSourceSpec `json:"spec"`
	// +optional
	Status VolumeUploadSourceStatus `json:"status,omitempty"`
}

// VolumeUploadSourceSpec defines specification for VolumeUploadSource
type VolumeUploadSourceSpec struct {
	// ContentType represents the type of the upload data (Kubevirt or archive)
	ContentType DataVolumeContentType `json:"contentType,omitempty"`
	// Preallocation controls whether storage for the target PVC should be allocated in advance.
	Preallocation *bool `json:"preallocation,omitempty"`
}

// VolumeUploadSourceStatus provides the most recently observed status of the VolumeUploadSource
type VolumeUploadSourceStatus struct {
}

// VolumeUploadSourceList provides the needed parameters to do request a list of Upload Sources from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VolumeUploadSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of DataSources
	Items []VolumeImportSource `json:"items"`
}

const (
	// VolumeImportSourceRef is import source for DataSourceRef for PVC
	VolumeImportSourceRef = "VolumeImportSource"
	// VolumeUploadSourceRef is upload source for DataSourceRef for PVC
	VolumeUploadSourceRef = "VolumeUploadSource"
	// VolumeCloneSourceRef is smart clone source for DataSourceRef for PVC
	VolumeCloneSourceRef = "VolumeCloneSource"
)

// VolumeCloneSource refers to a PVC/VolumeSnapshot of any storageclass/volumemode
// to be used as the source of a new PVC
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
type VolumeCloneSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumeCloneSourceSpec `json:"spec"`
}

// VolumeCloneSourceSpec defines the Spec field for VolumeCloneSource
type VolumeCloneSourceSpec struct {
	// Source is the src of the data to be cloned to the target PVC
	Source corev1.TypedLocalObjectReference `json:"source"`

	// Preallocation controls whether storage for the target PVC should be allocated in advance.
	// +optional
	Preallocation *bool `json:"preallocation,omitempty"`

	// PriorityClassName is the priorityclass for the claim
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`
}

// VolumeCloneSourceList provides the needed parameters to do request a list of VolumeCloneSources from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VolumeCloneSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	// Items provides a list of DataSources
	Items []VolumeCloneSource `json:"items"`
}

// this has to be here otherwise informer-gen doesn't recognize it
// see https://github.com/kubernetes/code-generator/issues/59
// +genclient:nonNamespaced

// CDI is the CDI Operator CRD
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=cdi;cdis,scope=Cluster
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
type CDI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CDISpec `json:"spec"`
	// +optional
	Status CDIStatus `json:"status"`
}

// CertConfig contains the tunables for TLS certificates
type CertConfig struct {
	// The requested 'duration' (i.e. lifetime) of the Certificate.
	Duration *metav1.Duration `json:"duration,omitempty"`

	// The amount of time before the currently issued certificate's `notAfter`
	// time that we will begin to attempt to renew the certificate.
	RenewBefore *metav1.Duration `json:"renewBefore,omitempty"`
}

// CDICertConfig has the CertConfigs for CDI
type CDICertConfig struct {
	// CA configuration
	// CA certs are kept in the CA bundle as long as they are valid
	CA *CertConfig `json:"ca,omitempty"`

	// Server configuration
	// Certs are rotated and discarded
	Server *CertConfig `json:"server,omitempty"`

	// Client configuration
	// Certs are rotated and discarded
	Client *CertConfig `json:"client,omitempty"`
}

// CDISpec defines our specification for the CDI installation
type CDISpec struct {
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// PullPolicy describes a policy for if/when to pull a container image
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" valid:"required"`
	// +kubebuilder:validation:Enum=RemoveWorkloads;BlockUninstallIfWorkloadsExist
	// CDIUninstallStrategy defines the state to leave CDI on uninstall
	UninstallStrategy *CDIUninstallStrategy `json:"uninstallStrategy,omitempty"`
	// Selectors and tolerations that should apply to cdi infrastructure components
	Infra ComponentConfig `json:"infra,omitempty"`
	// Restrict on which nodes CDI workload pods will be scheduled
	Workloads           sdkapi.NodePlacement `json:"workload,omitempty"`
	CustomizeComponents CustomizeComponents  `json:"customizeComponents,omitempty"`
	// Clone strategy override: should we use a host-assisted copy even if snapshots are available?
	// +kubebuilder:validation:Enum="copy";"snapshot";"csi-clone"
	CloneStrategyOverride *CDICloneStrategy `json:"cloneStrategyOverride,omitempty"`
	// CDIConfig at CDI level
	Config *CDIConfigSpec `json:"config,omitempty"`
	// certificate configuration
	CertConfig *CDICertConfig `json:"certConfig,omitempty"`
	// PriorityClass of the CDI control plane
	PriorityClass *CDIPriorityClass `json:"priorityClass,omitempty"`
}

// ComponentConfig defines the scheduling and replicas configuration for CDI components
type ComponentConfig struct {
	// NodePlacement describes scheduling configuration for specific CDI components
	sdkapi.NodePlacement `json:",inline"`
	// DeploymentReplicas set Replicas for cdi-deployment
	DeploymentReplicas *int32 `json:"deploymentReplicas,omitempty"`
	// ApiserverReplicas set Replicas for cdi-apiserver
	APIServerReplicas *int32 `json:"apiServerReplicas,omitempty"`
	// UploadproxyReplicas set Replicas for cdi-uploadproxy
	UploadProxyReplicas *int32 `json:"uploadProxyReplicas,omitempty"`
}

// CDIPriorityClass defines the priority class of the CDI control plane.
type CDIPriorityClass string

// CDICloneStrategy defines the preferred method for performing a CDI clone (override snapshot?)
type CDICloneStrategy string

const (
	// CloneStrategyHostAssisted specifies slower, host-assisted copy
	CloneStrategyHostAssisted CDICloneStrategy = "copy"

	// CloneStrategySnapshot specifies snapshot-based copying
	CloneStrategySnapshot CDICloneStrategy = "snapshot"

	// CloneStrategyCsiClone specifies csi volume clone based cloning
	CloneStrategyCsiClone CDICloneStrategy = "csi-clone"
)

// CustomizeComponents defines patches for components deployed by the CDI operator.
type CustomizeComponents struct {
	// +listType=atomic
	Patches []CustomizeComponentsPatch `json:"patches,omitempty"`

	// Configure the value used for deployment and daemonset resources
	Flags *Flags `json:"flags,omitempty"`
}

// Flags will create a patch that will replace all flags for the container's
// command field. The only flags that will be used are those define. There are no
// guarantees around forward/backward compatibility.  If set incorrectly this will
// cause the resource when rolled out to error until flags are updated.
type Flags struct {
	API         map[string]string `json:"api,omitempty"`
	Controller  map[string]string `json:"controller,omitempty"`
	UploadProxy map[string]string `json:"uploadProxy,omitempty"`
}

// CustomizeComponentsPatch defines a patch for some resource.
type CustomizeComponentsPatch struct {
	// +kubebuilder:validation:MinLength=1
	ResourceName string `json:"resourceName"`
	// +kubebuilder:validation:MinLength=1
	ResourceType string    `json:"resourceType"`
	Patch        string    `json:"patch"`
	Type         PatchType `json:"type"`
}

// PatchType defines the patch type.
type PatchType string

const (
	// JSONPatchType is a constant that represents the type of JSON patch.
	JSONPatchType PatchType = "json"
	// MergePatchType is a constant that represents the type of JSON Merge patch.
	MergePatchType PatchType = "merge"
	// StrategicMergePatchType is a constant that represents the type of Strategic Merge patch.
	StrategicMergePatchType PatchType = "strategic"
)

// DataImportCronSourceFormat defines the format of the DataImportCron-created disk image sources
type DataImportCronSourceFormat string

const (
	// DataImportCronSourceFormatSnapshot implies using a VolumeSnapshot as the resulting DataImportCron disk image source
	DataImportCronSourceFormatSnapshot DataImportCronSourceFormat = "snapshot"

	// DataImportCronSourceFormatPvc implies using a PVC as the resulting DataImportCron disk image source
	DataImportCronSourceFormatPvc DataImportCronSourceFormat = "pvc"
)

// CDIUninstallStrategy defines the state to leave CDI on uninstall
type CDIUninstallStrategy string

const (
	// CDIUninstallStrategyRemoveWorkloads specifies clean uninstall
	CDIUninstallStrategyRemoveWorkloads CDIUninstallStrategy = "RemoveWorkloads"

	// CDIUninstallStrategyBlockUninstallIfWorkloadsExist "leaves stuff around"
	CDIUninstallStrategyBlockUninstallIfWorkloadsExist CDIUninstallStrategy = "BlockUninstallIfWorkloadsExist"
)

// CDIPhase is the current phase of the CDI deployment
type CDIPhase string

// CDIStatus defines the status of the installation
type CDIStatus struct {
	sdkapi.Status `json:",inline"`
}

// CDIList provides the needed parameters to do request a list of CDIs from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CDIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of CDIs
	Items []CDI `json:"items"`
}

// this has to be here otherwise informer-gen doesn't recognize it
// see https://github.com/kubernetes/code-generator/issues/59
// +genclient:nonNamespaced

// CDIConfig provides a user configuration for CDI
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster
type CDIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CDIConfigSpec   `json:"spec"`
	Status CDIConfigStatus `json:"status,omitempty"`
}

// Percent is a string that can only be a value between [0,1)
// (Note: we actually rely on reconcile to reject invalid values)
// +kubebuilder:validation:Pattern=`^(0(?:\.\d{1,3})?|1)$`
type Percent string

// FilesystemOverhead defines the reserved size for PVCs with VolumeMode: Filesystem
type FilesystemOverhead struct {
	// Global is how much space of a Filesystem volume should be reserved for overhead. This value is used unless overridden by a more specific value (per storageClass)
	Global Percent `json:"global,omitempty"`
	// StorageClass specifies how much space of a Filesystem volume should be reserved for safety. The keys are the storageClass and the values are the overhead. This value overrides the global value
	StorageClass map[string]Percent `json:"storageClass,omitempty"`
}

// CDIConfigSpec defines specification for user configuration
type CDIConfigSpec struct {
	// Override the URL used when uploading to a DataVolume
	UploadProxyURLOverride *string `json:"uploadProxyURLOverride,omitempty"`
	// ImportProxy contains importer pod proxy configuration.
	// +optional
	ImportProxy *ImportProxy `json:"importProxy,omitempty"`
	// Override the storage class to used for scratch space during transfer operations. The scratch space storage class is determined in the following order: 1. value of scratchSpaceStorageClass, if that doesn't exist, use the default storage class, if there is no default storage class, use the storage class of the DataVolume, if no storage class specified, use no storage class for scratch space
	ScratchSpaceStorageClass *string `json:"scratchSpaceStorageClass,omitempty"`
	// ResourceRequirements describes the compute resource requirements.
	PodResourceRequirements *corev1.ResourceRequirements `json:"podResourceRequirements,omitempty"`
	// FeatureGates are a list of specific enabled feature gates
	FeatureGates []string `json:"featureGates,omitempty"`
	// FilesystemOverhead describes the space reserved for overhead when using Filesystem volumes. A value is between 0 and 1, if not defined it is 0.055 (5.5% overhead)
	FilesystemOverhead *FilesystemOverhead `json:"filesystemOverhead,omitempty"`
	// Preallocation controls whether storage for DataVolumes should be allocated in advance.
	Preallocation *bool `json:"preallocation,omitempty"`
	// InsecureRegistries is a list of TLS disabled registries
	InsecureRegistries []string `json:"insecureRegistries,omitempty"`
	// DataVolumeTTLSeconds is the time in seconds after DataVolume completion it can be garbage collected. Disabled by default.
	// Deprecated: Removed in v1.62.
	// +optional
	DataVolumeTTLSeconds *int32 `json:"dataVolumeTTLSeconds,omitempty"`
	// TLSSecurityProfile is used by operators to apply cluster-wide TLS security settings to operands.
	TLSSecurityProfile *TLSSecurityProfile `json:"tlsSecurityProfile,omitempty"`
	// The imagePullSecrets used to pull the container images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// LogVerbosity overrides the default verbosity level used to initialize loggers
	// +optional
	LogVerbosity *int32 `json:"logVerbosity,omitempty"`
}

// CDIConfigStatus provides the most recently observed status of the CDI Config resource
type CDIConfigStatus struct {
	// The calculated upload proxy URL
	UploadProxyURL *string `json:"uploadProxyURL,omitempty"`
	// UploadProxyCA is the certificate authority of the upload proxy
	UploadProxyCA *string `json:"uploadProxyCA,omitempty"`
	// ImportProxy contains importer pod proxy configuration.
	// +optional
	ImportProxy *ImportProxy `json:"importProxy,omitempty"`
	// The calculated storage class to be used for scratch space
	ScratchSpaceStorageClass string `json:"scratchSpaceStorageClass,omitempty"`
	// ResourceRequirements describes the compute resource requirements.
	DefaultPodResourceRequirements *corev1.ResourceRequirements `json:"defaultPodResourceRequirements,omitempty"`
	// FilesystemOverhead describes the space reserved for overhead when using Filesystem volumes. A percentage value is between 0 and 1
	FilesystemOverhead *FilesystemOverhead `json:"filesystemOverhead,omitempty"`
	// Preallocation controls whether storage for DataVolumes should be allocated in advance.
	Preallocation bool `json:"preallocation,omitempty"`
	// The imagePullSecrets used to pull the container images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// CDIConfigList provides the needed parameters to do request a list of CDIConfigs from the system
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CDIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items provides a list of CDIConfigs
	Items []CDIConfig `json:"items"`
}

// ImportProxy provides the information on how to configure the importer pod proxy.
type ImportProxy struct {
	// HTTPProxy is the URL http://<username>:<pswd>@<ip>:<port> of the import proxy for HTTP requests.  Empty means unset and will not result in the import pod env var.
	// +optional
	HTTPProxy *string `json:"HTTPProxy,omitempty"`
	// HTTPSProxy is the URL https://<username>:<pswd>@<ip>:<port> of the import proxy for HTTPS requests.  Empty means unset and will not result in the import pod env var.
	// +optional
	HTTPSProxy *string `json:"HTTPSProxy,omitempty"`
	// NoProxy is a comma-separated list of hostnames and/or CIDRs for which the proxy should not be used. Empty means unset and will not result in the import pod env var.
	// +optional
	NoProxy *string `json:"noProxy,omitempty"`
	// TrustedCAProxy is the name of a ConfigMap in the cdi namespace that contains a user-provided trusted certificate authority (CA) bundle.
	// The TrustedCAProxy ConfigMap is consumed by the DataImportCron controller for creating cronjobs, and by the import controller referring a copy of the ConfigMap in the import namespace.
	// Here is an example of the ConfigMap (in yaml):
	//
	// apiVersion: v1
	// kind: ConfigMap
	// metadata:
	//   name: my-ca-proxy-cm
	//   namespace: cdi
	// data:
	//   ca.pem: |
	//     -----BEGIN CERTIFICATE-----
	// 	   ... <base64 encoded cert> ...
	// 	   -----END CERTIFICATE-----
	// +optional
	TrustedCAProxy *string `json:"trustedCAProxy,omitempty"`
}
