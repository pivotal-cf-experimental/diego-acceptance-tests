package diego

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/diego-acceptance-tests/helpers/assets"
)

var _ = Describe("Application Lifecycle", func() {
	var appName string

	Context("Application with all buildpacks", func() {
		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		It("stages with auto detect and runs on diego", func() {
			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().HelloWorld, "--no-start"), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_STAGE_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_RUN_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), 2*CF_PUSH_TIMEOUT).Should(Exit(0)) // Double timeout to allow for all buildpacks.
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hello, world!"))
		})

		It("stages with a named buildpack and runs on diego", func() {
			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().HelloWorld, "--no-start", "-b", "ruby_buildpack"), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_STAGE_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_RUN_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hello, world!"))
		})

		It("stages with a zip buildpack and runs on diego", func() {
			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", ZIP_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_STAGE_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_RUN_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
		})

		It("stages with a git buildpack and runs on diego", func() {
			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", GIT_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_STAGE_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_RUN_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
		})
	})

	Context("Application with simple Null buildpack", func() {
		BeforeEach(func() {
			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", ZIP_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
		})

		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		Describe("An app staged with Diego and running on a DEA", func() {
			BeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, DIEGO_STAGE_BETA, "true")).Should(Exit(0))
				Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			It("exercises the app through its lifecycle", func() {
				By("verifying it's up")
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))

				By("stopping it")
				Eventually(cf.Cf("stop", appName)).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))

				By("starting it")
				Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
			})
		})

		Describe("An app staged on the DEA and running on Diego", func() {
			BeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, DIEGO_RUN_BETA, "true")).Should(Exit(0))
				Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			It("comes up", func() {
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
			})
		})
	})
})
