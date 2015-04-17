package diego

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/pivotal-cf-experimental/diego-acceptance-tests/helpers/assets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("FUSE", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	It("can mount a fuse endpoint", func() {
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Fuse, "-b", "ruby_buildpack", "--no-start"), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)
		Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))

		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("great success!"))
	})
})
