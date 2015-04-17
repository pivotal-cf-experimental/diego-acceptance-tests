package diego

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/pivotal-cf-experimental/diego-acceptance-tests/helpers/assets"
)

var _ = Describe("Buildpacks", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	It("stages with auto detect and runs on diego", func() {
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().HelloWorld, "--no-start"), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)
		Eventually(cf.Cf("start", appName), 2*CF_PUSH_TIMEOUT).Should(Exit(0)) // Double timeout to allow for all buildpacks.
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hello, world!"))
	})

	It("stages with a named buildpack and runs on diego", func() {
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().HelloWorld, "--no-start", "-b", "ruby_buildpack"), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)
		Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hello, world!"))
	})

	It("stages with a zip buildpack and runs on diego", func() {
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", ZIP_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)
		Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
	})

	It("stages with a git buildpack and runs on diego", func() {
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", GIT_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)
		session := cf.Cf("start", appName)
		Eventually(session, CF_PUSH_TIMEOUT).Should(Exit(0))
		Ω(session).Should(Say("LANG=en_US.UTF-8"))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
	})

	It("advertises the stack as CF_STACK when staging on diego", func() {
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", GIT_NULL_BUILDPACK, "-s", "cflinuxfs2"), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)
		session := cf.Cf("start", appName)
		Eventually(session, CF_PUSH_TIMEOUT).Should(Exit(0))
		Ω(session).Should(Say("CF_STACK=cflinuxfs2"))
	})
})
