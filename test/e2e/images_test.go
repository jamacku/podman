package integration

import (
	"fmt"
	"os"
	"sort"
	"strings"

	. "github.com/containers/podman/v3/test/utils"
	"github.com/docker/go-units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Podman images", func() {
	var (
		tempdir    string
		err        error
		podmanTest *PodmanTestIntegration
	)

	BeforeEach(func() {
		tempdir, err = CreateTempDirInTempDir()
		if err != nil {
			os.Exit(1)
		}
		podmanTest = PodmanTestCreate(tempdir)
		podmanTest.Setup()
		podmanTest.SeedImages()
	})

	AfterEach(func() {
		podmanTest.Cleanup()
		f := CurrentGinkgoTestDescription()
		processTestResult(f)

	})
	It("podman images", func() {
		session := podmanTest.Podman([]string{"images"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(BeNumerically(">", 2))
		Expect(session.LineInOutputStartsWith("quay.io/libpod/alpine")).To(BeTrue())
		Expect(session.LineInOutputStartsWith("quay.io/libpod/busybox")).To(BeTrue())
	})

	It("podman image List", func() {
		session := podmanTest.Podman([]string{"image", "list"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(BeNumerically(">", 2))
		Expect(session.LineInOutputStartsWith("quay.io/libpod/alpine")).To(BeTrue())
		Expect(session.LineInOutputStartsWith("quay.io/libpod/busybox")).To(BeTrue())
	})

	It("podman images with multiple tags", func() {
		// tag "docker.io/library/alpine:latest" to "foo:{a,b,c}"
		podmanTest.AddImageToRWStore(ALPINE)
		session := podmanTest.Podman([]string{"tag", ALPINE, "foo:a", "foo:b", "foo:c"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		// tag "foo:c" to "bar:{a,b}"
		session = podmanTest.Podman([]string{"tag", "foo:c", "bar:a", "bar:b"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		// check all previous and the newly tagged images
		session = podmanTest.Podman([]string{"images"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(session.LineInOutputContainsTag("quay.io/libpod/alpine", "latest")).To(BeTrue())
		Expect(session.LineInOutputContainsTag("quay.io/libpod/busybox", "latest")).To(BeTrue())
		Expect(session.LineInOutputContainsTag("localhost/foo", "a")).To(BeTrue())
		Expect(session.LineInOutputContainsTag("localhost/foo", "b")).To(BeTrue())
		Expect(session.LineInOutputContainsTag("localhost/foo", "c")).To(BeTrue())
		Expect(session.LineInOutputContainsTag("localhost/bar", "a")).To(BeTrue())
		Expect(session.LineInOutputContainsTag("localhost/bar", "b")).To(BeTrue())
		session = podmanTest.Podman([]string{"images", "-qn"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(BeNumerically("==", len(CACHE_IMAGES)))
	})

	It("podman images with digests", func() {
		session := podmanTest.Podman([]string{"images", "--digests"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(BeNumerically(">", 2))
		Expect(session.LineInOutputStartsWith("quay.io/libpod/alpine")).To(BeTrue())
		Expect(session.LineInOutputStartsWith("quay.io/libpod/busybox")).To(BeTrue())
	})

	It("podman empty images list in JSON format", func() {
		session := podmanTest.Podman([]string{"images", "--format=json", "not-existing-image"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(session.IsJSONOutputValid()).To(BeTrue())
	})

	It("podman images in JSON format", func() {
		session := podmanTest.Podman([]string{"images", "--format=json"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(session.IsJSONOutputValid()).To(BeTrue())
	})

	It("podman images in GO template format", func() {
		formatStr := "{{.ID}}\t{{.Created}}\t{{.CreatedAt}}\t{{.CreatedSince}}\t{{.CreatedTime}}"
		session := podmanTest.Podman([]string{"images", fmt.Sprintf("--format=%s", formatStr)})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
	})

	It("podman images with short options", func() {
		session := podmanTest.Podman([]string{"images", "-qn"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(BeNumerically(">", 1))
	})

	It("podman images filter by image name", func() {
		podmanTest.AddImageToRWStore(ALPINE)
		podmanTest.AddImageToRWStore(BB)

		session := podmanTest.Podman([]string{"images", "-q", ALPINE})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(Equal(1))

		session = podmanTest.Podman([]string{"tag", ALPINE, "foo:a"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		session = podmanTest.Podman([]string{"tag", BB, "foo:b"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		session = podmanTest.Podman([]string{"images", "-q", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(Equal(2))
	})

	It("podman images filter reference", func() {
		result := podmanTest.Podman([]string{"images", "-q", "-f", "reference=quay.io*"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
		Expect(len(result.OutputToStringArray())).To(Equal(7))

		retalpine := podmanTest.Podman([]string{"images", "-f", "reference=a*pine"})
		retalpine.WaitWithDefaultTimeout()
		Expect(retalpine).Should(Exit(0))
		Expect(len(retalpine.OutputToStringArray())).To(Equal(6))
		Expect(retalpine.LineInOutputContains("alpine")).To(BeTrue())

		retalpine = podmanTest.Podman([]string{"images", "-f", "reference=alpine"})
		retalpine.WaitWithDefaultTimeout()
		Expect(retalpine).Should(Exit(0))
		Expect(len(retalpine.OutputToStringArray())).To(Equal(6))
		Expect(retalpine.LineInOutputContains("alpine")).To(BeTrue())

		retnone := podmanTest.Podman([]string{"images", "-q", "-f", "reference=bogus"})
		retnone.WaitWithDefaultTimeout()
		Expect(retnone).Should(Exit(0))
		Expect(len(retnone.OutputToStringArray())).To(Equal(0))
	})

	It("podman images filter before image", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
RUN apk update && apk add strace
`
		podmanTest.BuildImage(dockerfile, "foobar.com/before:latest", "false")
		result := podmanTest.Podman([]string{"images", "-q", "-f", "before=foobar.com/before:latest"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
		Expect(len(result.OutputToStringArray()) >= 1).To(BeTrue())

	})

	It("podman images workingdir from  image", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
WORKDIR /test
`
		podmanTest.BuildImage(dockerfile, "foobar.com/workdir:latest", "false")
		result := podmanTest.Podman([]string{"run", "foobar.com/workdir:latest", "pwd"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
		Expect(result.OutputToString()).To(Equal("/test"))
	})

	It("podman images filter since image", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
`
		podmanTest.BuildImage(dockerfile, "foobar.com/before:latest", "false")
		result := podmanTest.Podman([]string{"images", "-q", "-f", "since=quay.io/libpod/alpine:latest"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
		Expect(len(result.OutputToStringArray())).To(Equal(9))
	})

	It("podman image list filter after image", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
`
		podmanTest.BuildImage(dockerfile, "foobar.com/before:latest", "false")
		result := podmanTest.Podman([]string{"image", "list", "-q", "-f", "after=quay.io/libpod/alpine:latest"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
		Expect(result.OutputToStringArray()).Should(HaveLen(9), "list filter output: %q", result.OutputToString())
	})

	It("podman images filter dangling", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
`
		podmanTest.BuildImage(dockerfile, "foobar.com/before:latest", "false")
		podmanTest.BuildImage(dockerfile, "foobar.com/before:latest", "false")
		result := podmanTest.Podman([]string{"images", "-q", "-f", "dangling=true"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0), "dangling image output: %q", result.OutputToString())
		Expect(result.OutputToStringArray()).Should(HaveLen(0), "dangling image output: %q", result.OutputToString())
	})

	It("podman pull by digest and list --all", func() {
		// Prevent regressing on issue #7651.
		digestPullAndList := func(noneTag bool) {
			session := podmanTest.Podman([]string{"pull", ALPINEAMD64DIGEST})
			session.WaitWithDefaultTimeout()
			Expect(session).Should(Exit(0))

			result := podmanTest.Podman([]string{"images", "--all", ALPINEAMD64DIGEST})
			result.WaitWithDefaultTimeout()
			Expect(result).Should(Exit(0))

			found, _ := result.GrepString("<none>")
			if noneTag {
				Expect(found).To(BeTrue())
			} else {
				Expect(found).To(BeFalse())
			}
		}
		// No "<none>" tag as tagged alpine instances should be present.
		session := podmanTest.Podman([]string{"pull", ALPINELISTTAG})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		digestPullAndList(false)

		// Now remove all images, re-pull by digest and check for the "<none>" tag.
		session = podmanTest.Podman([]string{"rmi", "-af"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		digestPullAndList(true)
	})

	It("podman check for image with sha256: prefix", func() {
		session := podmanTest.Podman([]string{"inspect", "--format=json", ALPINE})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(session.IsJSONOutputValid()).To(BeTrue())
		imageData := session.InspectImageJSON()

		result := podmanTest.Podman([]string{"images", "sha256:" + imageData[0].ID})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
	})

	It("podman check for image with sha256: prefix", func() {
		session := podmanTest.Podman([]string{"image", "inspect", "--format=json", ALPINE})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(session.IsJSONOutputValid()).To(BeTrue())
		imageData := session.InspectImageJSON()

		result := podmanTest.Podman([]string{"image", "ls", fmt.Sprintf("sha256:%s", imageData[0].ID)})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
	})

	It("podman images sort by values", func() {
		sortValueTest := func(value string, result int, format string) []string {
			f := fmt.Sprintf("{{.%s}}", format)
			session := podmanTest.Podman([]string{"images", "--noheading", "--sort", value, "--format", f})
			session.WaitWithDefaultTimeout()
			Expect(session).Should(Exit(result))

			return session.OutputToStringArray()
		}

		sortedArr := sortValueTest("created", 0, "CreatedAt")
		Expect(sort.SliceIsSorted(sortedArr, func(i, j int) bool { return sortedArr[i] > sortedArr[j] })).To(BeTrue())

		sortedArr = sortValueTest("id", 0, "ID")
		Expect(sort.SliceIsSorted(sortedArr, func(i, j int) bool { return sortedArr[i] < sortedArr[j] })).To(BeTrue())

		sortedArr = sortValueTest("repository", 0, "Repository")
		Expect(sort.SliceIsSorted(sortedArr, func(i, j int) bool { return sortedArr[i] < sortedArr[j] })).To(BeTrue())

		sortedArr = sortValueTest("size", 0, "Size")
		Expect(sort.SliceIsSorted(sortedArr, func(i, j int) bool {
			size1, _ := units.FromHumanSize(sortedArr[i])
			size2, _ := units.FromHumanSize(sortedArr[j])
			return size1 < size2
		})).To(BeTrue())
		sortedArr = sortValueTest("tag", 0, "Tag")
		Expect(sort.SliceIsSorted(sortedArr,
			func(i, j int) bool { return sortedArr[i] < sortedArr[j] })).
			To(BeTrue())

		sortValueTest("badvalue", 125, "Tag")
		sortValueTest("id", 125, "badvalue")
	})

	It("test for issue #6670", func() {
		expected := podmanTest.Podman([]string{"images", "--sort", "created", "--format", "{{.ID}}", "-q"})
		expected.WaitWithDefaultTimeout()

		actual := podmanTest.Podman([]string{"images", "--sort", "created", "-q"})
		actual.WaitWithDefaultTimeout()
		Expect(expected.Out).Should(Equal(actual.Out))
	})

	It("podman images --all flag", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
RUN mkdir hello
RUN touch test.txt
ENV foo=bar
`
		podmanTest.BuildImage(dockerfile, "test", "true")
		session := podmanTest.Podman([]string{"images"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(Equal(len(CACHE_IMAGES) + 2))

		session2 := podmanTest.Podman([]string{"images", "--all"})
		session2.WaitWithDefaultTimeout()
		Expect(session2).Should(Exit(0))
		Expect(len(session2.OutputToStringArray())).To(Equal(len(CACHE_IMAGES) + 4))
	})

	It("podman images filter by label", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
LABEL version="1.0"
LABEL "com.example.vendor"="Example Vendor"
`
		podmanTest.BuildImage(dockerfile, "test", "true")
		session := podmanTest.Podman([]string{"images", "-f", "label=version=1.0"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(Equal(2))
	})

	It("podman with images with no layers", func() {
		dockerfile := strings.Join([]string{
			`FROM scratch`,
			`LABEL org.opencontainers.image.authors="<somefolks@example.org>"`,
			`LABEL org.opencontainers.image.created=2019-06-11T19:03:37Z`,
			`LABEL org.opencontainers.image.description="This is a test image"`,
			`LABEL org.opencontainers.image.title=test`,
			`LABEL org.opencontainers.image.vendor="Example.org"`,
			`LABEL org.opencontainers.image.version=1`,
		}, "\n")
		podmanTest.BuildImage(dockerfile, "foo", "true")

		session := podmanTest.Podman([]string{"images", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		output := session.OutputToString()
		Expect(output).To(Not(MatchRegexp("<missing>")))
		Expect(output).To(Not(MatchRegexp("error")))

		session = podmanTest.Podman([]string{"image", "tree", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		output = session.OutputToString()
		Expect(output).To(MatchRegexp("No Image Layers"))

		session = podmanTest.Podman([]string{"history", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		output = session.OutputToString()
		Expect(output).To(Not(MatchRegexp("error")))

		session = podmanTest.Podman([]string{"history", "--quiet", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		Expect(len(session.OutputToStringArray())).To(Equal(6))

		session = podmanTest.Podman([]string{"image", "list", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		output = session.OutputToString()
		Expect(output).To(Not(MatchRegexp("<missing>")))
		Expect(output).To(Not(MatchRegexp("error")))

		session = podmanTest.Podman([]string{"image", "list"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		output = session.OutputToString()
		Expect(output).To(Not(MatchRegexp("<missing>")))
		Expect(output).To(Not(MatchRegexp("error")))

		session = podmanTest.Podman([]string{"inspect", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		output = session.OutputToString()
		Expect(output).To(Not(MatchRegexp("<missing>")))
		Expect(output).To(Not(MatchRegexp("error")))

		session = podmanTest.Podman([]string{"inspect", "--format", "{{.RootFS.Layers}}", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		output = session.OutputToString()
		Expect(output).To(Equal("[]"))
	})

	It("podman images --filter readonly", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
`
		podmanTest.BuildImage(dockerfile, "foobar.com/before:latest", "false")
		result := podmanTest.Podman([]string{"images", "-f", "readonly=true"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))

		result1 := podmanTest.Podman([]string{"images", "--filter", "readonly=false"})
		result1.WaitWithDefaultTimeout()
		Expect(result1).Should(Exit(0))
		Expect(result.OutputToStringArray()).To(Not(Equal(result1.OutputToStringArray())))
	})

	It("podman image prune --filter", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
RUN > file
`
		dockerfile2 := `FROM quay.io/libpod/alpine:latest
RUN > file2
`
		podmanTest.BuildImageWithLabel(dockerfile, "foobar.com/workdir:latest", "false", "abc")
		podmanTest.BuildImageWithLabel(dockerfile2, "foobar.com/workdir:latest", "false", "xyz")
		// --force used to to avoid y/n question
		result := podmanTest.Podman([]string{"image", "prune", "--filter", "label=abc", "--force"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
		Expect(len(result.OutputToStringArray())).To(Equal(1))

		//check if really abc is removed
		result = podmanTest.Podman([]string{"image", "list", "--filter", "label=abc"})
		Expect(len(result.OutputToStringArray())).To(Equal(0))

	})

	It("podman builder prune", func() {
		dockerfile := `FROM quay.io/libpod/alpine:latest
RUN > file
`
		dockerfile2 := `FROM quay.io/libpod/alpine:latest
RUN > file2
`
		podmanTest.BuildImageWithLabel(dockerfile, "foobar.com/workdir:latest", "false", "abc")
		podmanTest.BuildImageWithLabel(dockerfile2, "foobar.com/workdir:latest", "false", "xyz")
		// --force used to to avoid y/n question
		result := podmanTest.Podman([]string{"builder", "prune", "--filter", "label=abc", "--force"})
		result.WaitWithDefaultTimeout()
		Expect(result).Should(Exit(0))
		Expect(len(result.OutputToStringArray())).To(Equal(1))

		//check if really abc is removed
		result = podmanTest.Podman([]string{"image", "list", "--filter", "label=abc"})
		Expect(len(result.OutputToStringArray())).To(Equal(0))

	})

})
