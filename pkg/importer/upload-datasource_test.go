package importer

import (
	"io"
	"os"
	"path/filepath"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

const (
	dvKubevirt = cdiv1.DataVolumeKubeVirt
	dvArchive  = cdiv1.DataVolumeArchive
)

var _ = Describe("Upload data source", func() {
	var (
		ud     *UploadDataSource
		tmpDir string
		err    error
	)

	BeforeEach(func() {
		tmpDir, err = os.MkdirTemp("", "scratch")
		Expect(err).NotTo(HaveOccurred())
		By("tmpDir: " + tmpDir)
	})

	AfterEach(func() {
		if ud != nil {
			ud.Close()
		}
		os.RemoveAll(tmpDir)
	})

	It("Info should return Error, when passed in an image that cannot be read", func() {
		// Don't need to defer close, since ud.Close will close the reader
		file, err := os.Open(filepath.Join(imageDir, "content.tar"))
		Expect(err).NotTo(HaveOccurred())
		err = file.Close()
		Expect(err).NotTo(HaveOccurred())
		ud = NewUploadDataSource(file, dvKubevirt)
		result, err := ud.Info()
		Expect(err).To(HaveOccurred())
		Expect(ProcessingPhaseError).To(Equal(result))
	})

	It("Info should return Transfer, when passed in a valid image", func() {
		// Don't need to defer close, since ud.Close will close the reader
		file, err := os.Open(cirrosFilePath)
		Expect(err).NotTo(HaveOccurred())
		ud = NewUploadDataSource(file, dvKubevirt)
		result, err := ud.Info()

		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferScratch).To(Equal(result))
	})

	It("Info should return TransferDataDir with archive content type", func() {
		// Don't need to defer close, since ud.Close will close the reader
		file, err := os.Open(tinyCoreTarFilePath)
		Expect(err).NotTo(HaveOccurred())
		ud = NewUploadDataSource(file, dvArchive)
		result, err := ud.Info()

		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferDataDir).To(Equal(result))
	})

	It("Info should return TransferData, when passed in a valid raw image", func() {
		// Don't need to defer close, since ud.Close will close the reader
		file, err := os.Open(tinyCoreFilePath)
		Expect(err).NotTo(HaveOccurred())
		ud = NewUploadDataSource(file, dvKubevirt)
		result, err := ud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferDataFile).To(Equal(result))
	})

	DescribeTable("calling transfer should", func(fileName string, dvContentType cdiv1.DataVolumeContentType, expectedPhase ProcessingPhase, scratchPath string, want []byte, wantErr bool) {
		if scratchPath == "" {
			scratchPath = tmpDir
		}
		sourceFile, err := os.Open(fileName)
		Expect(err).NotTo(HaveOccurred())

		ud = NewUploadDataSource(sourceFile, dvContentType)
		_, err = ud.Info()
		Expect(err).NotTo(HaveOccurred())
		nextPhase, err := ud.Transfer(scratchPath, false)
		Expect(nextPhase).To(Equal(expectedPhase))
		if nextPhase == ProcessingPhaseConvert {
			Expect(err).NotTo(HaveOccurred())
			file, err := os.Open(filepath.Join(scratchPath, tempFile))
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()
			fileStat, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(int64(len(want))).To(Equal(fileStat.Size()))
			resultBuffer, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(reflect.DeepEqual(resultBuffer, want)).To(BeTrue())
			Expect(file.Name()).To(Equal(ud.GetURL().String()))
		} else if wantErr {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("return Error with missing scratch space", cirrosFilePath, dvKubevirt, ProcessingPhaseError, "/imaninvalidpath", nil, true),
		Entry("return Convert with scratch space and valid qcow file", cirrosFilePath, dvKubevirt, ProcessingPhaseConvert, "", cirrosData, false),
		Entry("return Complete with archive content type and archive file ", archiveFilePath, dvArchive, ProcessingPhaseComplete, "", []byte{}, false),
	)

	It("Transfer should fail on reader error", func() {
		sourceFile, err := os.Open(cirrosFilePath)
		Expect(err).NotTo(HaveOccurred())

		ud = NewUploadDataSource(sourceFile, dvKubevirt)
		nextPhase, err := ud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferScratch).To(Equal(nextPhase))
		err = sourceFile.Close()
		Expect(err).NotTo(HaveOccurred())
		result, err := ud.Transfer(tmpDir, false)
		Expect(err).To(HaveOccurred())
		Expect(ProcessingPhaseError).To(Equal(result))
	})

	It("TransferFile should succeed when writing to valid file", func() {
		// Don't need to defer close, since ud.Close will close the reader
		sourceFile, err := os.Open(tinyCoreFilePath)
		Expect(err).NotTo(HaveOccurred())
		ud = NewUploadDataSource(sourceFile, dvKubevirt)
		result, err := ud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferDataFile).To(Equal(result))
		result, err = ud.TransferFile(filepath.Join(tmpDir, "file"), false)
		Expect(err).ToNot(HaveOccurred())
		Expect(ProcessingPhaseResize).To(Equal(result))
	})

	It("TransferFile should fail on streaming error", func() {
		// Don't need to defer close, since ud.Close will close the reader
		sourceFile, err := os.Open(tinyCoreFilePath)
		Expect(err).NotTo(HaveOccurred())
		ud = NewUploadDataSource(sourceFile, dvKubevirt)
		result, err := ud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferDataFile).To(Equal(result))
		result, err = ud.TransferFile("/invalidpath/invalidfile", false)
		Expect(err).To(HaveOccurred())
		Expect(ProcessingPhaseError).To(Equal(result))
	})

	It("Close with nil stream should not fail", func() {
		ud = NewUploadDataSource(nil, dvKubevirt)
		err := ud.Close()
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Async Upload data source", func() {
	var (
		aud    *AsyncUploadDataSource
		tmpDir string
		err    error
	)

	BeforeEach(func() {
		tmpDir, err = os.MkdirTemp("", "scratch")
		Expect(err).NotTo(HaveOccurred())
		By("tmpDir: " + tmpDir)
	})

	AfterEach(func() {
		if aud != nil {
			aud.Close()
		}
		os.RemoveAll(tmpDir)
	})

	It("Info should return Error, when passed in an image that cannot be read", func() {
		// Don't need to defer close, since ud.Close will close the reader
		file, err := os.Open(filepath.Join(imageDir, "content.tar"))
		Expect(err).NotTo(HaveOccurred())
		err = file.Close()
		Expect(err).NotTo(HaveOccurred())
		aud = NewAsyncUploadDataSource(file)
		result, err := aud.Info()
		Expect(err).To(HaveOccurred())
		Expect(ProcessingPhaseError).To(Equal(result))
	})

	It("Info should return Transfer, when passed in a valid image", func() {
		// Don't need to defer close, since ud.Close will close the reader
		file, err := os.Open(cirrosFilePath)
		Expect(err).NotTo(HaveOccurred())
		aud = NewAsyncUploadDataSource(file)
		result, err := aud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferScratch).To(Equal(result))
	})

	It("Info should return TransferData, when passed in a valid raw image", func() {
		// Don't need to defer close, since ud.Close will close the reader
		file, err := os.Open(tinyCoreFilePath)
		Expect(err).NotTo(HaveOccurred())
		aud = NewAsyncUploadDataSource(file)
		result, err := aud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferDataFile).To(Equal(result))
	})

	DescribeTable("calling transfer should", func(fileName, scratchPath string, want []byte, wantErr bool) {
		if scratchPath == "" {
			scratchPath = tmpDir
		}
		sourceFile, err := os.Open(fileName)
		Expect(err).NotTo(HaveOccurred())

		aud = NewAsyncUploadDataSource(sourceFile)
		nextPhase, err := aud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferScratch).To(Equal(nextPhase))
		result, err := aud.Transfer(scratchPath, false)
		if !wantErr {
			Expect(err).NotTo(HaveOccurred())
			Expect(ProcessingPhaseValidatePause).To(Equal(result))
			Expect(ProcessingPhaseConvert).To(Equal(aud.GetResumePhase()))
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("return Error with missing scratch space", cirrosFilePath, "/imaninvalidpath", nil, true),
		Entry("return Convert with scratch space and valid qcow file", cirrosFilePath, "", cirrosData, false),
	)

	It("Transfer should fail on reader error", func() {
		sourceFile, err := os.Open(cirrosFilePath)
		Expect(err).NotTo(HaveOccurred())

		aud = NewAsyncUploadDataSource(sourceFile)
		nextPhase, err := aud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferScratch).To(Equal(nextPhase))
		err = sourceFile.Close()
		Expect(err).NotTo(HaveOccurred())
		result, err := aud.Transfer(tmpDir, false)
		Expect(err).To(HaveOccurred())
		Expect(ProcessingPhaseError).To(Equal(result))
	})

	It("TransferFile should succeed when writing to valid file", func() {
		// Don't need to defer close, since ud.Close will close the reader
		sourceFile, err := os.Open(tinyCoreFilePath)
		Expect(err).NotTo(HaveOccurred())
		aud = NewAsyncUploadDataSource(sourceFile)
		result, err := aud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferDataFile).To(Equal(result))
		result, err = aud.TransferFile(filepath.Join(tmpDir, "file"), false)
		Expect(err).ToNot(HaveOccurred())
		Expect(ProcessingPhaseValidatePause).To(Equal(result))
		Expect(ProcessingPhaseResize).To(Equal(aud.GetResumePhase()))
	})

	It("TransferFile should fail on streaming error", func() {
		// Don't need to defer close, since ud.Close will close the reader
		sourceFile, err := os.Open(tinyCoreFilePath)
		Expect(err).NotTo(HaveOccurred())
		aud = NewAsyncUploadDataSource(sourceFile)
		result, err := aud.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(ProcessingPhaseTransferDataFile).To(Equal(result))
		result, err = aud.TransferFile("/invalidpath/invalidfile", false)
		Expect(err).To(HaveOccurred())
		Expect(ProcessingPhaseError).To(Equal(result))
	})

	It("Close with nil stream should not fail", func() {
		aud = NewAsyncUploadDataSource(nil)
		err := aud.Close()
		Expect(err).NotTo(HaveOccurred())
	})
})
