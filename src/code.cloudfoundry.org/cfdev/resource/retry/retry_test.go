package retry_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/resource/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Retry", func() {
	It("retries until success", func() {
		counter := 0
		fn := func() error {
			counter += 1
			if counter < 6 {
				return fmt.Errorf("failing")
			}
			return nil
		}
		retryFn := func(error) bool { return true }
		Expect(retry.Retry(fn, retryFn)).To(Succeed())
		Expect(counter).To(Equal(6))
	})

	It("returns error if retryFn returns false", func() {
		fn := func() error { return fmt.Errorf("failing") }
		retryFn := func(error) bool { return false }

		Expect(retry.Retry(fn, retryFn)).To(MatchError("failing"))
	})

	Describe("Retryable", func() {
		It("does not retry other errors", func() {
			counter := 0
			fn := func() error {
				counter++
				return fmt.Errorf("failing")
			}

			Expect(retry.Retry(fn, retry.Retryable(10))).To(MatchError("failing"))
			Expect(counter).To(Equal(1))
		})

		It("retries retyables a max number of times", func() {
			counter := 0
			fn := func() error {
				counter++
				return retry.WrapAsRetryable(fmt.Errorf("failing"))
			}

			Expect(retry.Retry(fn, retry.Retryable(10))).To(MatchError("failing"))
			Expect(counter).To(Equal(10))
		})
	})
})
