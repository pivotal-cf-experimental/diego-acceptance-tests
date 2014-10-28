package diego

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/diego-acceptance-tests/helpers/assets"
)

var _ = Describe("Application Lifecycle", func() {
	var appName string

	Context("Application with all buildpacks", func() {
		It("should staged and run on diego without problem", func() {
			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Dora, "--no-start"), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, "DIEGO_STAGE_BETA", "true")).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, "DIEGO_RUN_BETA", "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), 2*CF_PUSH_TIMEOUT).Should(Exit(0)) // Double timeout to allow for all buildpacks.
		})

		AfterEach(func() {
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})
	})

	Context("Application with simple Null buildpack", func() {
		BeforeEach(func() {
			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", DIEGO_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
		})

		AfterEach(func() {
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		describeLifeCycle := func() {
			Describe("app lifecycle", func() {
				It("excercises the app through its lifecycle", func() {
					By("verifying it's up")
					Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))

					By("stopping it")
					Eventually(cf.Cf("stop", appName)).Should(Exit(0))
					Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))

					By("starting it")
					Eventually(cf.Cf("start", appName)).Should(Exit(0))
					Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))

					By("updating it")
					Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().HelloWorld, "-b", "ruby_buildpack"), CF_PUSH_TIMEOUT).Should(Exit(0))
					Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hello, world!"))

					By("deleting it")
					Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
					Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))
				})
			})
		}

		Describe("An app staged with Diego and running on a DEA", func() {
			BeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_STAGE_BETA", "true")).Should(Exit(0))
				Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			describeLifeCycle()
		})

		Describe("An app both staged and run with Diego", func() {
			BeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_STAGE_BETA", "true")).Should(Exit(0))
				Eventually(cf.Cf("set-env", appName, "DIEGO_RUN_BETA", "true")).Should(Exit(0))
				Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			describeLifeCycle()
		})

		Describe("An existing DEA-based app being migrated to Diego (staging only)", func() {
			BeforeEach(func() {
				Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "-b", DEA_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
				Eventually(cf.Cf("set-env", appName, "DIEGO_STAGE_BETA", "true")).Should(Exit(0))
			})

			Context("After repushing the app with a Diego compatible buildpack", func() {
				var push *Session

				BeforeEach(func() {
					push = cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "-b", DIEGO_NULL_BUILDPACK)
					Eventually(push, CF_PUSH_TIMEOUT).Should(Exit(0))
				})

				describeLifeCycle()

				It("is restaged with Diego", func() {
					Expect(push).To(Say("Uploading droplet, artifacts cache..."))
				})
			})

			Context("After restaging the app without changing the buildpack", func() {
				It("fails to restage because Diego does not support git buildpacks", func() {
					restart := cf.Cf("restage", appName)
					Eventually(restart, CF_PUSH_TIMEOUT).Should(Exit(1))
					Expect(restart).To(Say("Staging error: cannot get instances since staging failed"))
				})
			})
		})

		Describe("An existing DEA-based app being migrated to Diego (staging & running)", func() {
			BeforeEach(func() {
				Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "-b", DEA_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))

				Eventually(cf.Cf("stop", appName)).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))

				Eventually(cf.Cf("set-env", appName, "DIEGO_STAGE_BETA", "true")).Should(Exit(0))
				Eventually(cf.Cf("set-env", appName, "DIEGO_RUN_BETA", "true")).Should(Exit(0))
				Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "-b", DIEGO_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			describeLifeCycle()
		})
	})
})
