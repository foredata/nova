package circuit

// State changes between Closed, Open, HalfOpen
// [Closed] -->- tripped ----> [Open]<-------+
//    ^                          |           ^
//    |                          v           |
//    +                          |      detect fail
//    |                          |           |
//    |                    cooling timeout   |
//    ^                          |           ^
//    |                          v           |
//    +--- detect succeed --<-[HalfOpen]-->--+
//
// The behaviors of each states:
// =================================================================================================
// |           | [Succeed]                  | [Fail or Timeout]       | [IsAllowed]                |
// |================================================================================================
// | [Closed]  | do nothing                 | if tripped, become Open | allow                      |
// |================================================================================================
// | [Open]    | do nothing                 | do nothing              | if cooling timeout, allow; |
// |           |                            |                         | else reject                |
// |================================================================================================
// |           |increase halfopenSuccess,   |                         | if detect timeout, allow;  |
// |[HalfOpen] |if(halfopenSuccess >=       | become Open             | else reject                |
// |           | defaultHalfOpenSuccesses)|                         |                            |
// |           |     become Closed          |                         |                            |
// =================================================================================================
type State uint8

func (s State) String() string {
	switch s {
	case Open:
		return "OPEN"
	case HalfOpen:
		return "HALFOPEN"
	case Closed:
		return "CLOSED"
	}
	return "INVALID"
}

// represents the state
const (
	Open     State = iota
	HalfOpen State = iota
	Closed   State = iota
)
