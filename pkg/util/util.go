package util

import (
	"bufio"
	"bytes"
	"crypto/md5" //nolint:gosec // This is not a security-sensitive use case
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"kubevirt.io/containerized-data-importer/pkg/common"
)

const (
	// DefaultAlignBlockSize is the alignment size we use to align disk images, its a multiple of all known hardware block sizes 512/4k/8k/32k/64k.
	DefaultAlignBlockSize = 1024 * 1024
)

// CountingReader is a reader that keeps track of how much has been read
type CountingReader struct {
	Reader  io.ReadCloser
	Current uint64
	Done    bool
}

// RandAlphaNum provides an implementation to generate a random alpha numeric string of the specified length
// This generator is not cryptographically secure.
//
//nolint:gosec // This is not a security-sensitive use case
func RandAlphaNum(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[r.Intn(len(letter))]
	}
	return string(b)
}

// GetNamespace returns the namespace the pod is executing in
func GetNamespace() string {
	return getNamespace("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
}

func getNamespace(path string) string {
	if data, err := os.ReadFile(path); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "cdi"
}

// ParseEnvVar provides a wrapper to attempt to fetch the specified env var
func ParseEnvVar(envVarName string, decode bool) (string, error) {
	value := os.Getenv(envVarName)
	if decode {
		v, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return "", errors.Errorf("error decoding environment variable %q", envVarName)
		}
		value = string(v)
	}
	return value, nil
}

// Read reads bytes from the stream and updates the prometheus clone_progress metric according to the progress.
func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.Current += uint64(n)
	r.Done = errors.Is(err, io.EOF)
	return n, err
}

// Close closes the stream
func (r *CountingReader) Close() error {
	return r.Reader.Close()
}

// MinQuantity calculates the minimum of two quantities.
func MinQuantity(availableSpace, imageSize *resource.Quantity) resource.Quantity {
	if imageSize.Cmp(*availableSpace) == 1 {
		return *availableSpace
	}
	return *imageSize
}

// UnArchiveTar unarchives a tar file and streams its files
// using the specified io.Reader to the specified destination.
func UnArchiveTar(reader io.Reader, destDir string) error {
	klog.V(1).Infof("begin untar to %s...\n", destDir)
	untar := exec.Command("/usr/bin/tar", "--preserve-permissions", "--no-same-owner", "-xvC", destDir)
	untar.Stdin = reader
	var outBuf, errBuf bytes.Buffer
	untar.Stdout = &outBuf
	untar.Stderr = &errBuf
	klog.V(1).Infof("running untar cmd: %v\n", untar.Args)
	err := untar.Start()
	if err != nil {
		return err
	}
	err = untar.Wait()
	if err != nil {
		klog.V(3).Infof("STDOUT\n%s\n", outBuf.String())
		klog.V(3).Infof("STDERR\n%s\n", errBuf.String())
		klog.Errorf("%s\n", err.Error())
		return err
	}
	return nil
}

// WriteTerminationMessage writes the passed in message to the default termination message file
func WriteTerminationMessage(message string) error {
	return WriteTerminationMessageToFile(common.PodTerminationMessageFile, message)
}

// WriteTerminationMessageToFile writes the passed in message to the passed in message file
func WriteTerminationMessageToFile(file, message string) error {
	message = strings.ReplaceAll(message, "\n", " ")
	// Only write the first line of the message.
	scanner := bufio.NewScanner(strings.NewReader(message))

	if scanner.Scan() {
		err := os.WriteFile(file, scanner.Bytes(), 0600)
		if err != nil {
			return errors.Wrap(err, "could not create termination message file")
		}
	}
	return nil
}

// RoundDown returns the number rounded down to the nearest multiple.
func RoundDown(number, multiple int64) int64 {
	return number / multiple * multiple
}

// RoundUp returns the number rounded up to the nearest multiple.
func RoundUp(number, multiple int64) int64 {
	partitions := math.Ceil(float64(number) / float64(multiple))
	return int64(partitions) * multiple
}

// MergeLabels copies source labels to destination (overwrites existing labels)
func MergeLabels(src, dest map[string]string) map[string]string {
	if dest == nil {
		dest = map[string]string{}
	}

	for k, v := range src {
		dest[k] = v
	}

	return dest
}

// GetRecommendedInstallerLabelsFromCr returns the recommended labels to set on CDI resources
func GetRecommendedInstallerLabelsFromCr(cr *cdiv1.CDI) map[string]string {
	labels := map[string]string{}

	// In non-standalone installs, we fetch labels that were set on the CDI CR by the installer
	for k, v := range cr.GetLabels() {
		if k == common.AppKubernetesPartOfLabel || k == common.AppKubernetesVersionLabel {
			labels[k] = v
		}
	}

	return labels
}

// SetRecommendedLabels sets the recommended labels on CDI resources (does not get rid of existing ones)
func SetRecommendedLabels(obj metav1.Object, installerLabels map[string]string, controllerName string) {
	staticLabels := map[string]string{
		common.AppKubernetesManagedByLabel: controllerName,
		common.AppKubernetesComponentLabel: "storage",
	}

	// Merge static & existing labels
	mergedLabels := MergeLabels(staticLabels, obj.GetLabels())
	// Add installer dynamic labels as well (/version, /part-of)
	mergedLabels = MergeLabels(installerLabels, mergedLabels)

	obj.SetLabels(mergedLabels)
}

// Md5sum calculates the md5sum of a given file.
// Do not use this for security-sensitive use cases.
//
//nolint:gosec // This is not a security-sensitive use case
func Md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)[:16]
	return hex.EncodeToString(hashInBytes), nil
}

// GetUsableSpace calculates usable space to use taking file system overhead into account
func GetUsableSpace(filesystemOverhead float64, availableSpace int64) int64 {
	// Reverse the overhead calculation
	spaceWithoutOverhead := int64(math.Ceil(float64(availableSpace) / (1 + filesystemOverhead)))
	// qemu-img will round up, making us use more than the usable space.
	// This later conflicts with image size validation.
	return RoundDown(spaceWithoutOverhead, DefaultAlignBlockSize)
}

// GetRequiredSpace calculates space required taking file system overhead into account
func GetRequiredSpace(filesystemOverhead float64, requestedSpace int64) int64 {
	// the `image` has to be aligned correctly, so the space requested has to be aligned to
	// next value that is a multiple of a block size
	alignedSize := RoundUp(requestedSpace, DefaultAlignBlockSize)

	// count overhead as a percentage of the whole/new size, including aligned image
	// and the space required by filesystem metadata
	spaceWithOverhead := int64(math.Ceil(float64(alignedSize) * (1 + filesystemOverhead)))
	return spaceWithOverhead
}

// ResolveVolumeMode returns the volume mode if set, otherwise defaults to file system mode
func ResolveVolumeMode(volumeMode *v1.PersistentVolumeMode) v1.PersistentVolumeMode {
	retVolumeMode := v1.PersistentVolumeFilesystem
	if volumeMode != nil && *volumeMode == v1.PersistentVolumeBlock {
		retVolumeMode = v1.PersistentVolumeBlock
	}
	return retVolumeMode
}

// CopyFile copies a file from one location to another.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// CopyDir copies a dir from one location to another.
func CopyDir(source string, dest string) error {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		src := filepath.Join(source, obj.Name())
		dst := filepath.Join(dest, obj.Name())

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(src, dst)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(src, dst)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return err
}
