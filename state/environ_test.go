package state_test

import (
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/state"
)

type EnvironSuite struct {
	ConnSuite
	env *state.Environment
}

var _ = Suite(&EnvironSuite{})

func (s *EnvironSuite) SetUpTest(c *C) {
	s.ConnSuite.SetUpTest(c)
	setUpEnvConfig(c)
	env, err := s.State.Environment()
	c.Assert(err, IsNil)
	s.env = env
}

func (s *EnvironSuite) TestEntityName(c *C) {
	expected := "environment-" + envConfig["name"].(string)
	c.Assert(s.env.EntityName(), Equals, expected)
}

func (s *EnvironSuite) TestAnnotatorForEnvironment(c *C) {
	testAnnotator(c, func() (state.Annotator, error) {
		return s.State.Environment()
	})
}