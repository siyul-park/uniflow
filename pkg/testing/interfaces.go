package testing

import "time"

// 개별 테스트 케이스를 정의합니다.
type TestCase interface {
	// 테스트 케이스의 이름을 반환합니다.
	GetName() string
	// 테스트 노드의 명세를 반환합니다.
	GetSpec() *TestNodeSpec
	// 각 테스트 메서드 실행 전에 호출됩니다.
	SetUp() *TestResult
	// 각 테스트 메서드 실행 후에 호출됩니다.
	TearDown() *TestResult
	// 전체 테스트 시작 전 훅을 실행합니다.
	RunBeforeAll() *TestResult
	// 전체 테스트 완료 후 훅을 실행합니다.
	RunAfterAll() *TestResult
	// 등록된 모든 테스트 메서드들을 반환합니다.
	GetTestMethods() map[string]TestMethod
	// 지정된 이름의 테스트 메서드를 실행합니다.
	RunTest(name string, payload TestPayload) *TestResult
}

// 관련된 테스트 케이스들의 모음을 정의합니다.
type TestSuite interface {
	// 전체 테스트 스위트 실행 전에 한 번 호출됩니다.
	SetUpSuite() error
	// 전체 테스트 스위트 실행 후에 한 번 호출됩니다.
	TearDownSuite() error
	// 이 스위트에 포함된 모든 테스트 케이스들을 반환합니다.
	GetTestCases() []TestCase
}

// 테스트 실행을 담당합니다.
type TestRunner interface {
	// 주어진 테스트 케이스나 테스트 스위트를 실행합니다.
	Run(test interface{}) (*TestResult, error)
	// 모든 등록된 테스트들을 실행합니다.
	RunAll() ([]*TestResult, error)
	// 새로운 테스트 케이스를 등록합니다.
	RegisterTest(test TestCase) error
}

// 테스트 실행 결과를 나타냅니다.
type TestResult struct {
	Name      string        // 테스트 이름
	Error     error         // 발생한 오류
	StartTime time.Time     // 시작 시간
	EndTime   time.Time     // 종료 시간
	Status    TestStatus    // 실행 상태
	Children  []*TestResult // 하위 테스트 결과들
}

// 테스트의 실행 상태를 나타냅니다.
type TestStatus int

const (
	StatusPassed  TestStatus = iota // 성공
	StatusFailed                    // 실패
	StatusSkipped                   // 건너뜀
	StatusError                     // 오류
)
