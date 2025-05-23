package importer

import (
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"

	"kubevirt.io/containerized-data-importer/pkg/common"
)

const (
	s3FolderSep = "/"
	httpScheme  = "http"
)

// S3Client is the interface to the used S3 client.
type S3Client interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

// may be overridden in tests
var newClientFunc = getS3Client

// S3DataSource is the struct containing the information needed to import from an S3 data source.
// Sequence of phases:
// 1. Info -> Transfer
// 2. Transfer -> Convert
type S3DataSource struct {
	// S3 end point
	ep *url.URL
	// User name
	accessKey string
	// Password
	secKey string
	// Reader
	s3Reader io.ReadCloser
	// stack of readers
	readers *FormatReaders
	// The image file in scratch space.
	url *url.URL
}

// NewS3DataSource creates a new instance of the S3DataSource
func NewS3DataSource(endpoint, accessKey, secKey string, certDir string) (*S3DataSource, error) {
	ep, err := ParseEndpoint(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse endpoint %q", endpoint)
	}
	s3Reader, err := createS3Reader(ep, accessKey, secKey, certDir)
	if err != nil {
		return nil, err
	}
	return &S3DataSource{
		ep:        ep,
		accessKey: accessKey,
		secKey:    secKey,
		s3Reader:  s3Reader,
	}, nil
}

// Info is called to get initial information about the data.
func (sd *S3DataSource) Info() (ProcessingPhase, error) {
	var err error
	sd.readers, err = NewFormatReaders(sd.s3Reader, uint64(0))
	if err != nil {
		klog.Errorf("Error creating readers: %v", err)
		return ProcessingPhaseError, err
	}
	if !sd.readers.Convert {
		// Downloading a raw file, we can write that directly to the target.
		return ProcessingPhaseTransferDataFile, nil
	}

	return ProcessingPhaseTransferScratch, nil
}

// Transfer is called to transfer the data from the source to a temporary location.
func (sd *S3DataSource) Transfer(path string, preallocation bool) (ProcessingPhase, error) {
	file := filepath.Join(path, tempFile)
	if err := CleanAll(file); err != nil {
		return ProcessingPhaseError, err
	}

	size, _ := GetAvailableSpace(path)
	if size <= int64(0) {
		//Path provided is invalid.
		return ProcessingPhaseError, ErrInvalidPath
	}

	_, _, err := StreamDataToFile(sd.readers.TopReader(), file, preallocation)
	if err != nil {
		return ProcessingPhaseError, err
	}
	// If streaming succeeded, then parsing the file into URL will also succeed, no need to check error status
	sd.url, _ = url.Parse(file)
	return ProcessingPhaseConvert, nil
}

// TransferFile is called to transfer the data from the source to the passed in file.
func (sd *S3DataSource) TransferFile(fileName string, preallocation bool) (ProcessingPhase, error) {
	if err := CleanAll(fileName); err != nil {
		return ProcessingPhaseError, err
	}

	_, _, err := StreamDataToFile(sd.readers.TopReader(), fileName, preallocation)
	if err != nil {
		return ProcessingPhaseError, err
	}
	return ProcessingPhaseResize, nil
}

// GetURL returns the url that the data processor can use when converting the data.
func (sd *S3DataSource) GetURL() *url.URL {
	return sd.url
}

// GetTerminationMessage returns data to be serialized and used as the termination message of the importer.
func (sd *S3DataSource) GetTerminationMessage() *common.TerminationMessage {
	return nil
}

// Close closes any readers or other open resources.
func (sd *S3DataSource) Close() error {
	var err error
	if sd.readers != nil {
		err = sd.readers.Close()
	}
	return err
}

func createS3Reader(ep *url.URL, accessKey, secKey string, certDir string) (io.ReadCloser, error) {
	klog.V(3).Infoln("Using S3 client to get data")

	endpoint := ep.Host
	urlScheme := ep.Scheme
	klog.Infof("Endpoint %s", endpoint)
	path := strings.Trim(ep.Path, "/")
	bucket, object := extractBucketAndObject(path)

	klog.V(1).Infof("bucket %s", bucket)
	klog.V(1).Infof("object %s", object)
	svc, err := newClientFunc(endpoint, accessKey, secKey, certDir, urlScheme)
	if err != nil {
		return nil, errors.Wrapf(err, "could not build s3 client for %q", ep.Host)
	}

	objInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	objOutput, err := svc.GetObject(objInput)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get s3 object: \"%s/%s\"", bucket, object)
	}
	objectReader := objOutput.Body
	return objectReader, nil
}

func getS3Client(endpoint, accessKey, secKey string, certDir string, urlScheme string) (S3Client, error) {
	// Adding certs using CustomCABundle will overwrite the SystemCerts, so we opt by creating a custom HTTPClient
	httpClient, err := createHTTPClient(certDir)

	if err != nil {
		return nil, errors.Wrap(err, "Error creating http client for s3")
	}

	creds := credentials.NewStaticCredentials(accessKey, secKey, "")
	region := extractRegion(endpoint)
	disableSSL := false
	// Disable SSL for http endpoint. This should cause the s3 client to create http requests.
	if urlScheme == httpScheme {
		disableSSL = true
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      creds,
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       httpClient,
		DisableSSL:       &disableSSL,
	},
	)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	return svc, nil
}

func extractRegion(s string) string {
	var region string
	r, _ := regexp.Compile(`s3\.(.+)\.amazonaws\.com`)
	if matches := r.FindStringSubmatch(s); matches != nil {
		region = matches[1]
	} else {
		region = strings.Split(s, ".")[0]
	}

	return region
}

func extractBucketAndObject(s string) (string, string) {
	pathSplit := strings.Split(s, s3FolderSep)
	bucket := pathSplit[0]
	object := strings.Join(pathSplit[1:], s3FolderSep)
	return bucket, object
}
