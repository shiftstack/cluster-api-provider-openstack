/*
Copyright 2024 The Kubernetes Authors.

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

package network

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	orcv1alpha1 "github.com/k-orc/openstack-resource-controller/api/v1alpha1"

	"sigs.k8s.io/cluster-api-provider-openstack/pkg/clients/mock"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/scope"
)

type serveFileHandler string

func (s serveFileHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	http.ServeFile(resp, req, string(s))
}

const (
	rawImage = "raw.img"
	bz2Image = "raw.img.bz2"
	gzImage  = "raw.img.gz"
	xzImage  = "raw.img.xz"
)

type hashes struct {
	md5, sha1, sha256, sha512 string
}

// These values were generated by running (md5|sha1|sha256|sha512)sum against the testdata fixtures and recording the result here.
var fixtureHashes = map[string]hashes{
	rawImage: {
		md5:    "b6d81b360a5672d80c27430f39153e2c",
		sha1:   "3b71f43ff30f4b15b5cd85dd9e95ebc7e84eb5a3",
		sha256: "30e14955ebf1352266dc2ff8067e68104607e750abb9d3b36582b8af909fcb58",
		sha512: "d6292685b380e338e025b3415a90fe8f9d39a46e7bdba8cb78c50a338cefca741f69e4e46411c32de1afdedfb268e579a51f81ff85e56f55b0ee7c33fe8c25c9",
	},
	bz2Image: {
		md5:    "7769b1d98c72af733a5bc64730f1792b",
		sha1:   "e5f220ecf5b7a9e8956533a401a6dc884f9ca628",
		sha256: "b13f5e5969992b8dfa39808e8107d2232c4ecfea3011b2f23e3c17e7b8ec6e94",
		sha512: "bc789073b702aa3b67e838edf3402d083972065f955d889cb844e8480a03c8e08882dd8a103c94ed3f45e6b4e2c0f862facedb581a798d79262b09ee905bbaa6",
	},
	gzImage: {
		md5:    "d4ae7671832b29a7a58989e125619bc9",
		sha1:   "ed1bbe66b1a760ffe94229df22b3d06908003495",
		sha256: "453b059e88a2ced8f02bab81099881f64fb9870abb846c6bd5cb3d54753b63af",
		sha512: "27674bfa3fdd0133324399cda83fa7e561a1e597930524a9b370ecc3010b19764944698f17ef1235016781b8ce825079fe5ce9a96be5fb2217a20d45cc2994cc",
	},
	xzImage: {
		md5:    "f4c28650618347a7f6a7266479345bce",
		sha1:   "2e93ef446ea49935f08aeb77aa818312b51d49ef",
		sha256: "d4b8d7ffff1a03cffd5ed6e955ae14ad91dda417d5a0d315cfe7e71f853a3885",
		sha512: "acfe0cd5e41142be81c3b52c2d5fd8953300e35f83a713625ee160a3a69457df2c9af8b039ef1e97b5d47bb2dc5e0d6a688e752cf415b474d38f712de75826ba",
	},
}

type uploadTestOpts struct {
	downloadHash       *orcv1alpha1.ImageHash
	imageHashAlgorithm *orcv1alpha1.ImageHashAlgorithm
	compression        *orcv1alpha1.ImageCompression
	optionalUploadCall bool
}

type uploadTestOpt func(opts *uploadTestOpts)

func downloadHash(algorithm orcv1alpha1.ImageHashAlgorithm, value string) uploadTestOpt {
	return func(opts *uploadTestOpts) {
		opts.downloadHash = &orcv1alpha1.ImageHash{
			Algorithm: algorithm,
			Value:     value,
		}
	}
}

func imageHashAlgorithm(algorithm orcv1alpha1.ImageHashAlgorithm) uploadTestOpt {
	return func(opts *uploadTestOpts) {
		opts.imageHashAlgorithm = &algorithm
	}
}

func compression(compression orcv1alpha1.ImageCompression) uploadTestOpt {
	return func(opts *uploadTestOpts) {
		opts.compression = &compression
	}
}

func optionalUpload() uploadTestOpt {
	return func(opts *uploadTestOpts) {
		opts.optionalUploadCall = true
	}
}

var _ = Describe("Upload tests", Ordered, func() {
	var (
		ctx           context.Context
		namespace     *corev1.Namespace
		fileServeAddr string
		reconciler    *orcNetworkReconciler
		mockCtrl      *gomock.Controller
		scopeFactory  *scope.MockScopeFactory
	)

	BeforeAll(func() {
		httpCtx, done := context.WithCancel(context.Background())

		serveMux := http.NewServeMux()
		serveMux.Handle(`/raw.img`, serveFileHandler(`testdata/raw.img`))
		serveMux.Handle(`/raw.img.bz2`, serveFileHandler(`testdata/raw.img.bz2`))
		serveMux.Handle(`/raw.img.gz`, serveFileHandler(`testdata/raw.img.gz`))
		serveMux.Handle(`/raw.img.xz`, serveFileHandler(`testdata/raw.img.xz`))
		server := &http.Server{
			Handler:           serveMux,
			ReadHeaderTimeout: time.Second,
		}

		DeferCleanup(func() {
			server.Close()
			done()
		})

		// Allow the listener to find us any free port on localhost
		listenConfig := &net.ListenConfig{}
		listener, err := listenConfig.Listen(httpCtx, "tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred(), "create http listen socket")
		fileServeAddr = listener.Addr().String()

		go func() {
			_ = server.Serve(listener)
		}()
	})

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.TODO(), GinkgoLogr)

		// Create the namespace
		namespace = &corev1.Namespace{}
		namespace.SetGenerateName("test-")
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed(), "create namespace")
		DeferCleanup(func() {
			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed(), "delete namespace")
		})

		mockCtrl = gomock.NewController(GinkgoT())
		scopeFactory = scope.NewMockScopeFactory(mockCtrl, "")
		reconciler = &orcNetworkReconciler{
			client:       k8sClient,
			scopeFactory: scopeFactory,
			// NOTE(mdbooth): I have no idea what 1024 means here, or if it is a sensible value 🤷
			recorder: record.NewFakeRecorder(1024),
		}
	})

	uploadTest := func(imageName string, aopts ...uploadTestOpt) error {
		opts := uploadTestOpts{}
		for _, opt := range aopts {
			opt(&opts)
		}

		orcImage := &orcv1alpha1.Image{}
		orcImage.SetName("test-image")
		orcImage.SetNamespace(namespace.GetName())
		orcImage.Spec = orcv1alpha1.ImageSpec{
			Resource: &orcv1alpha1.ImageResourceSpec{
				Content: &orcv1alpha1.ImageContent{
					ContainerFormat: orcv1alpha1.ImageContainerFormatBare,
					DiskFormat:      orcv1alpha1.ImageDiskFormatRaw,
					SourceType:      orcv1alpha1.ImageSourceTypeURL,
					SourceURL: &orcv1alpha1.ImageContentSourceURL{
						URL:          "http://" + fileServeAddr + "/" + imageName,
						Decompress:   opts.compression,
						DownloadHash: opts.downloadHash,
					},
				},
			},
			CloudCredentialsRef: orcv1alpha1.CloudCredentialsReference{
				SecretName: "my-secret",
				CloudName:  "my-cloud",
			},
		}

		Expect(k8sClient.Create(ctx, orcImage)).To(Succeed(), "create ORC Image")
		DeferCleanup(func() {
			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed(), "delete ORC Image")
		})

		glanceImage := &images.Image{
			ID:              "image-id",
			Name:            "test-image",
			Status:          images.ImageStatusQueued,
			ContainerFormat: "bare",
			DiskFormat:      "raw",
		}

		imageClient := mock.NewMockImageClient(mockCtrl)
		recorder := imageClient.EXPECT()
		uploadDataMock := recorder.UploadData(ctx, glanceImage.ID, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, data io.Reader) error {
			_, err := io.ReadAll(data)
			return err
		})
		if opts.optionalUploadCall {
			uploadDataMock.MaxTimes(1)
		}

		return reconciler.uploadImageContent(ctx, orcImage, imageClient, glanceImage)
	}

	DescribeTable("should download a raw image",
		func(imageName string, opts ...uploadTestOpt) {
			Expect(uploadTest(imageName, opts...)).To(Succeed(), "uploadImageContent()")
		},
		Entry("without verification", rawImage),
		Entry("with md5 verification", rawImage, downloadHash(orcv1alpha1.ImageHashAlgorithmMD5, fixtureHashes[rawImage].md5)),
		Entry("with sha1 verification", rawImage, downloadHash(orcv1alpha1.ImageHashAlgorithmSHA1, fixtureHashes[rawImage].sha1)),
		Entry("with sha256 verification", rawImage, downloadHash(orcv1alpha1.ImageHashAlgorithmSHA256, fixtureHashes[rawImage].sha256)),
		Entry("with sha512 verification", rawImage, downloadHash(orcv1alpha1.ImageHashAlgorithmSHA512, fixtureHashes[rawImage].sha512)),
	)

	DescribeTable("should download and verify image compressed with",
		func(imageName string, opts ...uploadTestOpt) {
			opts = append(opts,
				downloadHash(orcv1alpha1.ImageHashAlgorithmSHA256, fixtureHashes[imageName].sha256),
				imageHashAlgorithm(orcv1alpha1.ImageHashAlgorithmSHA512),
			)

			err := uploadTest(imageName, opts...)
			Expect(err).To(Succeed(), "uploadImageContent()")
		},
		Entry("bz", bz2Image, compression(orcv1alpha1.ImageCompressionBZ2)),
		Entry("gz", gzImage, compression(orcv1alpha1.ImageCompressionGZ)),
		Entry("xz", xzImage, compression(orcv1alpha1.ImageCompressionXZ)),
	)

	DescribeTable("should fail download when verification fails with algorithm",
		func(algorithm orcv1alpha1.ImageHashAlgorithm) {
			err := uploadTest(rawImage, downloadHash(algorithm, "aaaa"))
			Expect(err).NotTo(Succeed(), "uploadImageContent()")
		},
		Entry("md5", orcv1alpha1.ImageHashAlgorithmMD5),
		Entry("sha1", orcv1alpha1.ImageHashAlgorithmSHA1),
		Entry("sha256", orcv1alpha1.ImageHashAlgorithmSHA256),
		Entry("sha512", orcv1alpha1.ImageHashAlgorithmSHA512),
	)

	DescribeTable("should fail to uncompress invalid",
		func(algorithm orcv1alpha1.ImageCompression) {
			// Some compression algorithms read data and return an
			// error on creation, others don't read data until data
			// is requested. This means that some compression
			// algorithms fail before calling UploadData, and some
			// fail during it.
			err := uploadTest(rawImage, compression(algorithm), optionalUpload())
			Expect(err).NotTo(Succeed(), "uploadImageContent()")
		},
		Entry("bz2", orcv1alpha1.ImageCompressionBZ2),
		Entry("gz", orcv1alpha1.ImageCompressionGZ),
		Entry("xz", orcv1alpha1.ImageCompressionXZ),
	)
})
