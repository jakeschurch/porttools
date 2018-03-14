package porttools

type BacktestEngine struct {
	Portfolio *Portfolio
	Benchmark *Index
	Log       *PerformanceLog
	Strategy  *Strategy
}

type PortManager interface {
	OrderLogic() bool // REVIEW: rename ConstraintLogic?
}

// TODO: Finish method signature
type Algorithm interface {
	EntryLogic() bool
	ExitLogic() bool
}

// Strategy ... TODO
type Strategy struct {
	algos []Algorithm
}

// PerformanceLog conducts performance analysis.
type PerformanceLog struct {
	Closed    PositionSlice
	orders    Queue
	benchmark *Index
}

// TODO: ExecuteStrategy method

//
// import (
// 	"encoding/csv"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"
// )

// option to throw out securities that aren't needed
// only see different trades -> simulate market that has traders doing different trades and what is the aggregate position look like
