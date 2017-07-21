package tests

import (
	"errors"
	"fmt"
	"path"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
)

// Done is part of the Ginkgo DSL
type Done ginkgo.Done

// Benchmarker is part of the Ginkgo DSL
type Benchmarker ginkgo.Benchmarker

// GinkgoWriter is part of the Ginkgo DSL
var GinkgoWriter = ginkgo.GinkgoWriter

// GinkgoParallelNode is part of the Ginkgo DSL
var GinkgoParallelNode = ginkgo.GinkgoParallelNode

// GinkgoT is part of the Ginkgo DSL
var GinkgoT = ginkgo.GinkgoT

// CurrentGinkgoTestDescription is part of the Ginkgo DSL
var CurrentGinkgoTestDescription = ginkgo.CurrentGinkgoTestDescription

// RunSpecs is part of the Ginkgo DSL
var RunSpecs = ginkgo.RunSpecs

// RunSpecsWithDefaultAndCustomReporters is part of the Ginkgo DSL
var RunSpecsWithDefaultAndCustomReporters = ginkgo.RunSpecsWithDefaultAndCustomReporters

// RunSpecsWithCustomReporters is part of the Ginkgo DSL
var RunSpecsWithCustomReporters = ginkgo.RunSpecsWithCustomReporters

// Skip is part of the Ginkgo DSL
var Skip = ginkgo.Skip

// Fail is part of the Ginkgo DSL
var Fail = ginkgo.Fail

// GinkgoRecover is part of the Ginkgo DSL
var GinkgoRecover = ginkgo.GinkgoRecover

// Describe is part of the Ginkgo DSL
var Describe = ginkgo.Describe

// FDescribe is part of the Ginkgo DSL
var FDescribe = ginkgo.FDescribe

// PDescribe is part of the Ginkgo DSL
var PDescribe = ginkgo.PDescribe

// XDescribe is part of the Ginkgo DSL
var XDescribe = ginkgo.XDescribe

// Context is part of the Ginkgo DSL
var Context = ginkgo.Context

// FContext is part of the Ginkgo DSL
var FContext = ginkgo.FContext

// PContext is part of the Ginkgo DSL
var PContext = ginkgo.PContext

// XContext is part of the Ginkgo DSL
var XContext = ginkgo.XContext

// It is part of the Ginkgo DSL
var It = ginkgo.It

// FIt is part of the Ginkgo DSL
var FIt = ginkgo.FIt

// PIt is part of the Ginkgo DSL
var PIt = ginkgo.PIt

// XIt is part of the Ginkgo DSL
var XIt = ginkgo.XIt

// Specify is part of the Ginkgo DSL
var Specify = ginkgo.Specify

// FSpecify is part of the Ginkgo DSL
var FSpecify = ginkgo.FSpecify

// PSpecify is part of the Ginkgo DSL
var PSpecify = ginkgo.PSpecify

// XSpecify is part of the Ginkgo DSL
var XSpecify = ginkgo.XSpecify

// By is part of the Ginkgo DSL
var By = ginkgo.By

// Measure is part of the Ginkgo DSL
var Measure = ginkgo.Measure

// FMeasure is part of the Ginkgo DSL
var FMeasure = ginkgo.FMeasure

// PMeasure is part of the Ginkgo DSL
var PMeasure = ginkgo.PMeasure

// XMeasure is part of the Ginkgo DSL
var XMeasure = ginkgo.XMeasure

// BeforeSuite is part of the Ginkgo DSL
var BeforeSuite = ginkgo.BeforeSuite

// AfterSuite is part of the Ginkgo DSL
var AfterSuite = ginkgo.AfterSuite

// SynchronizedBeforeSuite is part of the Ginkgo DSL
var SynchronizedBeforeSuite = ginkgo.SynchronizedBeforeSuite

// SynchronizedAfterSuite is part of the Ginkgo DSL
var SynchronizedAfterSuite = ginkgo.SynchronizedAfterSuite

// BeforeEach is part of the Ginkgo DSL
var BeforeEach = ginkgo.BeforeEach

// JustBeforeEach is part of the Ginkgo DSL
var JustBeforeEach = ginkgo.JustBeforeEach

// AfterEach is part of the Ginkgo DSL
var AfterEach = ginkgo.AfterEach

// Declarations for Gomega DSL

// RegisterFailHandler is part of the Ginkgo DSL
var RegisterFailHandler = gomega.RegisterFailHandler

// RegisterTestingT is part of the Ginkgo DSL
var RegisterTestingT = gomega.RegisterTestingT

// InterceptGomegaFailures is part of the Ginkgo DSL
var InterceptGomegaFailures = gomega.InterceptGomegaFailures

// Ω is part of the Ginkgo DSL
var Ω = gomega.Ω

// Expect is part of the Ginkgo DSL
var Expect = gomega.Expect

// ExpectWithOffset is part of the Ginkgo DSL
var ExpectWithOffset = gomega.ExpectWithOffset

// Eventually is part of the Ginkgo DSL
var Eventually = gomega.Eventually

// EventuallyWithOffset is part of the Ginkgo DSL
var EventuallyWithOffset = gomega.EventuallyWithOffset

// Consistently is part of the Ginkgo DSL
var Consistently = gomega.Consistently

// ConsistentlyWithOffset is part of the Ginkgo DSL
var ConsistentlyWithOffset = gomega.ConsistentlyWithOffset

// SetDefaultEventuallyTimeout is part of the Ginkgo DSL
var SetDefaultEventuallyTimeout = gomega.SetDefaultEventuallyTimeout

// SetDefaultEventuallyPollingInterval is part of the Ginkgo DSL
var SetDefaultEventuallyPollingInterval = gomega.SetDefaultEventuallyPollingInterval

// SetDefaultConsistentlyDuration is part of the Ginkgo DSL
var SetDefaultConsistentlyDuration = gomega.SetDefaultConsistentlyDuration

// SetDefaultConsistentlyPollingInterval is part of the Ginkgo DSL
var SetDefaultConsistentlyPollingInterval = gomega.SetDefaultConsistentlyPollingInterval

// Declarations for Gomega Matchers

// Equal is part of the Ginkgo DSL
var Equal = gomega.Equal

// BeEquivalentTo is part of the Ginkgo DSL
var BeEquivalentTo = gomega.BeEquivalentTo

// BeIdenticalTo is part of the Ginkgo DSL
var BeIdenticalTo = gomega.BeIdenticalTo

// BeNil is part of the Ginkgo DSL
var BeNil = gomega.BeNil

// BeTrue is part of the Ginkgo DSL
var BeTrue = gomega.BeTrue

// BeFalse is part of the Ginkgo DSL
var BeFalse = gomega.BeFalse

// HaveOccurred is part of the Ginkgo DSL
var HaveOccurred = gomega.HaveOccurred

// Succeed is part of the Ginkgo DSL
var Succeed = gomega.Succeed

// MatchError is part of the Ginkgo DSL
var MatchError = gomega.MatchError

// BeClosed is part of the Ginkgo DSL
var BeClosed = gomega.BeClosed

// Receive is part of the Ginkgo DSL
var Receive = gomega.Receive

// BeSent is part of the Ginkgo DSL
var BeSent = gomega.BeSent

// MatchRegexp is part of the Ginkgo DSL
var MatchRegexp = gomega.MatchRegexp

// ContainSubstring is part of the Ginkgo DSL
var ContainSubstring = gomega.ContainSubstring

// HavePrefix is part of the Ginkgo DSL
var HavePrefix = gomega.HavePrefix

// HaveSuffix is part of the Ginkgo DSL
var HaveSuffix = gomega.HaveSuffix

// MatchJSON is part of the Ginkgo DSL
var MatchJSON = gomega.MatchJSON

// BeEmpty is part of the Ginkgo DSL
var BeEmpty = gomega.BeEmpty

// HaveLen is part of the Ginkgo DSL
var HaveLen = gomega.HaveLen

// HaveCap is part of the Ginkgo DSL
var HaveCap = gomega.HaveCap

// BeZero is part of the Ginkgo DSL
var BeZero = gomega.BeZero

// ContainElement is part of the Ginkgo DSL
var ContainElement = gomega.ContainElement

// ConsistOf is part of the Ginkgo DSL
var ConsistOf = gomega.ConsistOf

// HaveKey is part of the Ginkgo DSL
var HaveKey = gomega.HaveKey

// HaveKeyWithValue is part of the Ginkgo DSL
var HaveKeyWithValue = gomega.HaveKeyWithValue

// BeNumerically is part of the Ginkgo DSL
var BeNumerically = gomega.BeNumerically

// BeTemporally is part of the Ginkgo DSL
var BeTemporally = gomega.BeTemporally

// BeAssignableToTypeOf is part of the Ginkgo DSL
var BeAssignableToTypeOf = gomega.BeAssignableToTypeOf

// Panic is part of the Ginkgo DSL
var Panic = gomega.Panic

// BeAnExistingFile is part of the Ginkgo DSL
var BeAnExistingFile = gomega.BeAnExistingFile

// BeARegularFile is part of the Ginkgo DSL
var BeARegularFile = gomega.BeARegularFile

// BeADirectory is part of the Ginkgo DSL
var BeADirectory = gomega.BeADirectory

// And is part of the Ginkgo DSL
var And = gomega.And

// SatisfyAll is part of the Ginkgo DSL
var SatisfyAll = gomega.SatisfyAll

// Or is part of the Ginkgo DSL
var Or = gomega.Or

// SatisfyAny is part of the Ginkgo DSL
var SatisfyAny = gomega.SatisfyAny

// Not is part of the Ginkgo DSL
var Not = gomega.Not

// WithTransform is part of the Ginkgo DSL
var WithTransform = gomega.WithTransform

type pathMatcher struct {
	expected string
}

func (pm *pathMatcher) Match(actual interface{}) (bool, error) {
	szA, ok := actual.(string)
	if !ok {
		return false, errors.New("pathMatcher expects one or more strings")
	}
	return pm.expected == szA, nil
}
func (pm *pathMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nto be equal to\n\t%#v", actual, pm.expected)
}
func (pm *pathMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nnot to be equal to\n\t%#v", actual, pm.expected)
}

// Σ is a custom Ginkgo matcher.
func Σ(paths ...string) gomegaTypes.GomegaMatcher {
	return &pathMatcher{path.Join(paths...)}
}
