package sumoratinglifecycle

import (
	"errors"
	"fmt"

	"github.com/jph5396/sumomodel"
)

type (
	//RikishiData contains the data for rikishi
	// and their rating.
	RikishiData struct {
		Rikishi sumomodel.Rikishi
		Rating  float32
	}

	// BoutResult contains data representing the result of a bout.
	BoutResult struct {
		BashoID int
		Day     int
		BoutNum int
		east    RikishiBoutResult
		west    RikishiBoutResult
		winner  string
	}

	//RikishiBoutResult represents one rikishi's peformance in a bout.
	// including the id, name, score, and change since the last bout.
	RikishiBoutResult struct {
		RikishiID int
		Name      string
		Score     float32
		Change    float32
	}

	//Sumocycle contains data and lifecycle hooks to calculate sumo
	// ratings.
	Sumocycle struct {
		Basho       int
		Day         int
		RikishiData map[int]RikishiData
		BoutList    []sumomodel.Bout
		resultlist  []BoutResult

		preday    func(*Sumocycle)
		prebout   func(*sumomodel.Bout, int)
		calculate func(float32, float32, bool) float32
		postbout  func(BoutResult)
		postday   func(Sumocycle)
	}
)

// NewSumocycle creates a new sumo cycle struct. All functions are set to nil
func NewSumocycle(basho int, day int, rikishi map[int]RikishiData, boutlist []sumomodel.Bout) Sumocycle {
	cycle := Sumocycle{
		Basho:       basho,
		Day:         day,
		RikishiData: rikishi,
		BoutList:    boutlist,
	}

	return cycle
}

// BeforeDay sets the function that is executed at the beginning
// of the basho before any calculations begin
func (s *Sumocycle) BeforeDay(f func(*Sumocycle)) {
	s.preday = f

}

// BeforeBout set the function that will execute right before a
// bout
func (s *Sumocycle) BeforeBout(f func(*sumomodel.Bout, int)) {
	s.prebout = f
}

// AfterBout set the function that will execute after a bout.
func (s *Sumocycle) AfterBout(f func(BoutResult)) {
	s.postbout = f
}

// AfterDay set the function that will execute after all bout calculations
// have been completed. It is a good time to save data if desired.
func (s *Sumocycle) AfterDay(f func(Sumocycle)) {
	s.postday = f

}

// Calculation set the function that calculates the rating.
// the provided function accepts two floats and a bool.
// the first float represents the current rikishi and the second represents
// their opponent.
func (s *Sumocycle) Calculation(f func(float32, float32, bool) float32) {
	s.calculate = f
}

// Begin checks if all required functions on the Sumocycle object are set.
// if yes, it begins the process. If no, it returns an error.
func (s *Sumocycle) Begin() error {
	err := s.validate()
	if err != nil {
		return err
	}

	if s.preday != nil {
		s.preday(s)
	}

	for i, bout := range s.BoutList {
		if s.prebout != nil {
			s.prebout(&bout, i)
		}
		// Gather Rikishi Data
		east, ok := s.RikishiData[bout.EastRikishiID]
		if !ok {
			return fmt.Errorf("Rikishi with id %v was not provided but appears in a bout", bout.EastRikishiID)
		}

		west, ok := s.RikishiData[bout.WestRikishiID]
		if !ok {
			return fmt.Errorf("Rikishi with id %v was not provided but appears in a bout", bout.WestRikishiID)
		}

		eastNewScore := s.calculate(east.Rating, west.Rating, bout.EastWin)
		westNewScore := s.calculate(west.Rating, east.Rating, bout.WestWin)

		// build
		eastBoutResult := RikishiBoutResult{
			RikishiID: east.Rikishi.Id,
			Name:      east.Rikishi.Name,
			Score:     eastNewScore,
			Change:    eastNewScore - east.Rating,
		}

		westBoutResult := RikishiBoutResult{
			RikishiID: west.Rikishi.Id,
			Name:      west.Rikishi.Name,
			Score:     westNewScore,
			Change:    westNewScore - west.Rating,
		}

		newBoutResult := BoutResult{
			BashoID: bout.BashoID,
			Day:     bout.Day,
			BoutNum: bout.Boutnum,
			east:    eastBoutResult,
			west:    westBoutResult,
		}

		//Execute the postbout function if one exists.
		if s.postbout != nil {
			s.postbout(newBoutResult)
		}

		east.Rating = eastNewScore
		west.Rating = westNewScore
		s.RikishiData[east.Rikishi.Id] = east
		s.RikishiData[west.Rikishi.Id] = west

		s.resultlist = append(s.resultlist, newBoutResult)
	}

	if s.postday != nil {
		s.postday(*s)
	}

	return nil
}

func (s Sumocycle) validate() error {
	if s.calculate == nil {
		return errors.New("no calculate function set")
	}

	if len(s.RikishiData) == 0 {
		return errors.New("No Rikishi data provided")
	}
	if len(s.BoutList) == 0 {
		return errors.New("No bout list provided")
	}
	return nil
}
