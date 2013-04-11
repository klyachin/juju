package main

import (
	"fmt"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/bzr"
	"launchpad.net/juju-core/charm"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/testing"
	"os"
	"time"
)

// Sadly, this is a very slow test suite, heavily dominated by calls to bzr.

type PublishSuite struct {
	testing.LoggingSuite
	testing.HTTPSuite

	dir        string
	oldBaseURL string
	branch     *bzr.Branch
}

var _ = Suite(&PublishSuite{})

func touch(c *C, filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	c.Assert(err, IsNil)
	f.Close()
}

func addMeta(c *C, branch *bzr.Branch, meta string) {
	if meta == "" {
		meta = "name: wordpress\nsummary: Some summary\ndescription: Some description.\n"
	}
	f, err := os.Create(branch.Join("metadata.yaml"))
	c.Assert(err, IsNil)
	_, err = f.Write([]byte(meta))
	f.Close()
	c.Assert(err, IsNil)
	err = branch.Add("metadata.yaml")
	c.Assert(err, IsNil)
	err = branch.Commit("Added metadata.yaml.")
	c.Assert(err, IsNil)
}

func (s *PublishSuite) runPublish(c *C, args ...string) (*cmd.Context, error) {
	return testing.RunCommandInDir(c, &PublishCommand{}, args, s.dir)
}

const pollDelay = 100 * time.Millisecond

func (s *PublishSuite) SetUpSuite(c *C) {
	s.LoggingSuite.SetUpSuite(c)
	s.HTTPSuite.SetUpSuite(c)

	s.oldBaseURL = charm.Store.BaseURL
	charm.Store.BaseURL = s.URL("")
}

func (s *PublishSuite) TearDownSuite(c *C) {
	s.LoggingSuite.TearDownSuite(c)
	s.HTTPSuite.TearDownSuite(c)

	charm.Store.BaseURL = s.oldBaseURL
}

func (s *PublishSuite) SetUpTest(c *C) {
	s.LoggingSuite.SetUpTest(c)
	s.HTTPSuite.SetUpTest(c)

	s.dir = c.MkDir()
	s.branch = bzr.New(s.dir)
	err := s.branch.Init()
	c.Assert(err, IsNil)
}

func (s *PublishSuite) TearDownTest(c *C) {
	s.HTTPSuite.TearDownTest(c)
	s.LoggingSuite.TearDownTest(c)
}

func (s *PublishSuite) TestNoBranch(c *C) {
	dir := c.MkDir()
	_, err := testing.RunCommandInDir(c, &PublishCommand{}, []string{"cs:precise/wordpress"}, dir)
	c.Assert(err, ErrorMatches, fmt.Sprintf("not a charm branch: %s", dir))
}

func (s *PublishSuite) TestEmpty(c *C) {
	_, err := s.runPublish(c, "cs:precise/wordpress")
	c.Assert(err, ErrorMatches, `cannot obtain local digest: branch has no content`)
}

func (s *PublishSuite) TestFrom(c *C) {
	_, err := testing.RunCommandInDir(c, &PublishCommand{}, []string{"--from", s.dir, "cs:precise/wordpress"}, c.MkDir())
	c.Assert(err, ErrorMatches, `cannot obtain local digest: branch has no content`)
}

func (s *PublishSuite) TestMissingSeries(c *C) {
	_, err := s.runPublish(c, "cs:wordpress")
	c.Assert(err, ErrorMatches, `cannot infer charm URL for "cs:wordpress": no series provided`)
}

func (s *PublishSuite) TestNotClean(c *C) {
	touch(c, s.branch.Join("file"))
	_, err := s.runPublish(c, "cs:precise/wordpress")
	c.Assert(err, ErrorMatches, `branch is not clean \(bzr status\)`)
}

func (s *PublishSuite) TestNoPushLocation(c *C) {
	addMeta(c, s.branch, "")
	_, err := s.runPublish(c)
	c.Assert(err, ErrorMatches, `no charm URL provided and cannot infer from current directory \(no push location\)`)
}

func (s *PublishSuite) TestUnknownPushLocation(c *C) {
	addMeta(c, s.branch, "")
	err := s.branch.Push(&bzr.PushAttr{Location: c.MkDir() + "/foo", Remember: true})
	c.Assert(err, IsNil)
	_, err = s.runPublish(c)
	c.Assert(err, ErrorMatches, `cannot infer charm URL from branch location: ".*/foo"`)
}

func (s *PublishSuite) TestWrongRepository(c *C) {
	addMeta(c, s.branch, "")
	_, err := s.runPublish(c, "local:precise/wordpress")
	c.Assert(err, ErrorMatches, "charm URL must reference the juju charm store")
}

func (s *PublishSuite) TestInferURL(c *C) {
	addMeta(c, s.branch, "")

	cmd := &PublishCommand{}
	cmd.ChangePushLocation(func(location string) string {
		c.Assert(location, Equals, "lp:charms/precise/wordpress")
		c.SucceedNow()
		panic("unreachable")
	})

	_, err := testing.RunCommandInDir(c, cmd, []string{"precise/wordpress"}, s.dir)
	c.Assert(err, IsNil)
	c.Fatal("shouldn't get here; location closure didn't run?")
}

func (s *PublishSuite) TestBrokenCharm(c *C) {
	addMeta(c, s.branch, "name: wordpress\nsummary: Some summary\n")
	_, err := s.runPublish(c, "cs:precise/wordpress")
	c.Assert(err, ErrorMatches, "metadata: description: expected string, got nothing")
}

func (s *PublishSuite) TestWrongName(c *C) {
	addMeta(c, s.branch, "")
	_, err := s.runPublish(c, "cs:precise/mysql")
	c.Assert(err, ErrorMatches, `charm name in metadata must match name in URL: "wordpress" != "mysql"`)
}

func (s *PublishSuite) TestPreExistingPublished(c *C) {
	addMeta(c, s.branch, "")

	// Pretend the store has seen the digest before, and it has succeeded.
	digest, err := s.branch.RevisionId()
	c.Assert(err, IsNil)
	body := `{"cs:precise/wordpress": {"kind": "published", "digest": %q, "revision": 42}}`
	testing.Server.Response(200, nil, []byte(fmt.Sprintf(body, digest)))

	ctx, err := s.runPublish(c, "cs:precise/wordpress")
	c.Assert(err, IsNil)
	c.Assert(testing.Stdout(ctx), Equals, "cs:precise/wordpress-42\n")

	req := testing.Server.WaitRequest()
	c.Assert(req.URL.Path, Equals, "/charm-event")
	c.Assert(req.Form.Get("charms"), Equals, "cs:precise/wordpress@"+digest)
}

func (s *PublishSuite) TestPreExistingPublishedEdge(c *C) {
	addMeta(c, s.branch, "")

	// If it doesn't find the right digest on the first try, it asks again for
	// any digest at all to keep the tip in mind. There's a small chance that
	// on the second request the tip has changed and matches the digest we're
	// looking for, in which case we have the answer already.
	digest, err := s.branch.RevisionId()
	c.Assert(err, IsNil)
	var body string
	body = `{"cs:precise/wordpress": {"errors": ["entry not found"]}}`
	testing.Server.Response(200, nil, []byte(body))
	body = `{"cs:precise/wordpress": {"kind": "published", "digest": %q, "revision": 42}}`
	testing.Server.Response(200, nil, []byte(fmt.Sprintf(body, digest)))

	ctx, err := s.runPublish(c, "cs:precise/wordpress")
	c.Assert(err, IsNil)
	c.Assert(testing.Stdout(ctx), Equals, "cs:precise/wordpress-42\n")

	req := testing.Server.WaitRequest()
	c.Assert(req.URL.Path, Equals, "/charm-event")
	c.Assert(req.Form.Get("charms"), Equals, "cs:precise/wordpress@"+digest)

	req = testing.Server.WaitRequest()
	c.Assert(req.URL.Path, Equals, "/charm-event")
	c.Assert(req.Form.Get("charms"), Equals, "cs:precise/wordpress")
}

func (s *PublishSuite) TestPreExistingPublishError(c *C) {
	addMeta(c, s.branch, "")

	// Pretend the store has seen the digest before, and it has failed.
	digest, err := s.branch.RevisionId()
	c.Assert(err, IsNil)
	body := `{"cs:precise/wordpress": {"kind": "publish-error", "digest": %q, "errors": ["an error"]}}`
	testing.Server.Response(200, nil, []byte(fmt.Sprintf(body, digest)))

	_, err = s.runPublish(c, "cs:precise/wordpress")
	c.Assert(err, ErrorMatches, "charm could not be published: an error")

	req := testing.Server.WaitRequest()
	c.Assert(req.URL.Path, Equals, "/charm-event")
	c.Assert(req.Form.Get("charms"), Equals, "cs:precise/wordpress@"+digest)
}

func (s *PublishSuite) TestFullPublish(c *C) {
	addMeta(c, s.branch, "")

	digest, err := s.branch.RevisionId()
	c.Assert(err, IsNil)

	pushBranch := bzr.New(c.MkDir())
	err = pushBranch.Init()
	c.Assert(err, IsNil)

	cmd := &PublishCommand{}
	cmd.ChangePushLocation(func(location string) string {
		c.Assert(location, Equals, "lp:~user/charms/precise/wordpress/trunk")
		return pushBranch.Location()
	})
	cmd.SetPollDelay(pollDelay)

	var body string

	// The local digest isn't found.
	body = `{"cs:~user/precise/wordpress": {"kind": "", "errors": ["entry not found"]}}`
	testing.Server.Response(200, nil, []byte(body))

	// But the charm exists with an arbitrary non-matching digest.
	body = `{"cs:~user/precise/wordpress": {"kind": "published", "digest": "other-digest"}}`
	testing.Server.Response(200, nil, []byte(body))

	// After the branch is pushed we fake the publishing delay.
	body = `{"cs:~user/precise/wordpress": {"kind": "published", "digest": "other-digest"}}`
	testing.Server.Response(200, nil, []byte(body))

	// And finally report success.
	body = `{"cs:~user/precise/wordpress": {"kind": "published", "digest": %q, "revision": 42}}`
	testing.Server.Response(200, nil, []byte(fmt.Sprintf(body, digest)))

	ctx, err := testing.RunCommandInDir(c, cmd, []string{"cs:~user/precise/wordpress"}, s.dir)
	c.Assert(err, IsNil)
	c.Assert(testing.Stdout(ctx), Equals, "cs:~user/precise/wordpress-42\n")

	// Ensure the branch was actually pushed.
	pushDigest, err := pushBranch.RevisionId()
	c.Assert(err, IsNil)
	c.Assert(pushDigest, Equals, digest)

	// And that all the requests were sent with the proper data.
	req := testing.Server.WaitRequest()
	c.Assert(req.URL.Path, Equals, "/charm-event")
	c.Assert(req.Form.Get("charms"), Equals, "cs:~user/precise/wordpress@"+digest)

	for i := 0; i < 3; i++ {
		// The second request grabs tip to see the current state, and the
		// following requests are done after pushing to see when it changes.
		req = testing.Server.WaitRequest()
		c.Assert(req.URL.Path, Equals, "/charm-event")
		c.Assert(req.Form.Get("charms"), Equals, "cs:~user/precise/wordpress")
	}
}

func (s *PublishSuite) TestFullPublishError(c *C) {
	addMeta(c, s.branch, "")

	digest, err := s.branch.RevisionId()
	c.Assert(err, IsNil)

	pushBranch := bzr.New(c.MkDir())
	err = pushBranch.Init()
	c.Assert(err, IsNil)

	cmd := &PublishCommand{}
	cmd.ChangePushLocation(func(location string) string {
		c.Assert(location, Equals, "lp:~user/charms/precise/wordpress/trunk")
		return pushBranch.Location()
	})
	cmd.SetPollDelay(pollDelay)

	var body string

	// The local digest isn't found.
	body = `{"cs:~user/precise/wordpress": {"kind": "", "errors": ["entry not found"]}}`
	testing.Server.Response(200, nil, []byte(body))

	// And tip isn't found either, meaning the charm was never published.
	testing.Server.Response(200, nil, []byte(body))

	// After the branch is pushed we fake the publishing delay.
	testing.Server.Response(200, nil, []byte(body))

	// And finally report success.
	body = `{"cs:~user/precise/wordpress": {"kind": "published", "digest": %q, "revision": 42}}`
	testing.Server.Response(200, nil, []byte(fmt.Sprintf(body, digest)))

	ctx, err := testing.RunCommandInDir(c, cmd, []string{"cs:~user/precise/wordpress"}, s.dir)
	c.Assert(err, IsNil)
	c.Assert(testing.Stdout(ctx), Equals, "cs:~user/precise/wordpress-42\n")

	// Ensure the branch was actually pushed.
	pushDigest, err := pushBranch.RevisionId()
	c.Assert(err, IsNil)
	c.Assert(pushDigest, Equals, digest)

	// And that all the requests were sent with the proper data.
	req := testing.Server.WaitRequest()
	c.Assert(req.URL.Path, Equals, "/charm-event")
	c.Assert(req.Form.Get("charms"), Equals, "cs:~user/precise/wordpress@"+digest)

	for i := 0; i < 3; i++ {
		// The second request grabs tip to see the current state, and the
		// following requests are done after pushing to see when it changes.
		req = testing.Server.WaitRequest()
		c.Assert(req.URL.Path, Equals, "/charm-event")
		c.Assert(req.Form.Get("charms"), Equals, "cs:~user/precise/wordpress")
	}
}

func (s *PublishSuite) TestFullPublishRace(c *C) {
	addMeta(c, s.branch, "")

	digest, err := s.branch.RevisionId()
	c.Assert(err, IsNil)

	pushBranch := bzr.New(c.MkDir())
	err = pushBranch.Init()
	c.Assert(err, IsNil)

	cmd := &PublishCommand{}
	cmd.ChangePushLocation(func(location string) string {
		c.Assert(location, Equals, "lp:~user/charms/precise/wordpress/trunk")
		return pushBranch.Location()
	})
	cmd.SetPollDelay(pollDelay)

	var body string

	// The local digest isn't found.
	body = `{"cs:~user/precise/wordpress": {"kind": "", "errors": ["entry not found"]}}`
	testing.Server.Response(200, nil, []byte(body))

	// And tip isn't found either, meaning the charm was never published.
	testing.Server.Response(200, nil, []byte(body))

	// After the branch is pushed we fake the publishing delay.
	testing.Server.Response(200, nil, []byte(body))

	// But, surprisingly, the digest changed to something else entirely.
	body = `{"cs:~user/precise/wordpress": {"kind": "published", "digest": "surprising-digest", "revision": 42}}`
	testing.Server.Response(200, nil, []byte(body))

	_, err = testing.RunCommandInDir(c, cmd, []string{"cs:~user/precise/wordpress"}, s.dir)
	c.Assert(err, ErrorMatches, `charm changed but not to local charm digest; publishing race\?`)

	// Ensure the branch was actually pushed.
	pushDigest, err := pushBranch.RevisionId()
	c.Assert(err, IsNil)
	c.Assert(pushDigest, Equals, digest)

	// And that all the requests were sent with the proper data.
	req := testing.Server.WaitRequest()
	c.Assert(req.URL.Path, Equals, "/charm-event")
	c.Assert(req.Form.Get("charms"), Equals, "cs:~user/precise/wordpress@"+digest)

	for i := 0; i < 3; i++ {
		// The second request grabs tip to see the current state, and the
		// following requests are done after pushing to see when it changes.
		req = testing.Server.WaitRequest()
		c.Assert(req.URL.Path, Equals, "/charm-event")
		c.Assert(req.Form.Get("charms"), Equals, "cs:~user/precise/wordpress")
	}
}