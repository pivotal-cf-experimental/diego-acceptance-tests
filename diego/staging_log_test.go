package diego

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/pivotal-cf-experimental/diego-acceptance-tests/helpers/assets"
)

var _ = Describe("An application being staged with Diego", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	It("has its staging log streamed during a push", func() {
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Dora, "-b", "ruby_buildpack", "--no-start"), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)

		start := cf.Cf("start", appName)

		Eventually(start, CF_PUSH_TIMEOUT).Should(Exit(0))
		Expect(start).Should(Say("Downloaded app package"))
		Expect(start).Should(Say(`Downloaded ruby_buildpack`))
		Expect(start).Should(Say(`Staging\.\.\.`))
		Expect(start).Should(Say("Staging complete"))
		Expect(start).Should(Say("Uploading complete"))
	})
})
