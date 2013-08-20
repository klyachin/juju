// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing

// Provides a TestDataSuite which creates and provides http access to sample simplestreams metadata.

import (
	"fmt"
	"net/http"

	gc "launchpad.net/gocheck"

	"launchpad.net/juju-core/environs/jujutest"
	"launchpad.net/juju-core/environs/simplestreams"
	"launchpad.net/juju-core/testing"
)

var imageData = map[string]string{
	"/streams/v1/index.json": `
		{
		 "index": {
		  "com.ubuntu.cloud:released:precise": {
		   "updated": "Wed, 01 May 2013 13:31:26 +0000",
		   "clouds": [
			{
			 "region": "us-east-1",
			 "endpoint": "https://ec2.us-east-1.amazonaws.com"
			}
		   ],
		   "cloudname": "aws",
		   "datatype": "image-ids",
		   "format": "products:1.0",
		   "products": [
			"com.ubuntu.cloud:server:12.04:amd64",
			"com.ubuntu.cloud:server:12.04:arm"
		   ],
		   "path": "streams/v1/image_metadata.json"
		  },
		  "com.ubuntu.cloud:released:raring": {
		   "updated": "Wed, 01 May 2013 13:31:26 +0000",
		   "clouds": [
			{
			 "region": "us-east-1",
			 "endpoint": "https://ec2.us-east-1.amazonaws.com"
			}
		   ],
		   "cloudname": "aws",
		   "datatype": "image-ids",
		   "format": "products:1.0",
		   "products": [
			"com.ubuntu.cloud:server:13.04:amd64"
		   ],
		   "path": "streams/v1/raring_metadata.json"
		  },
		  "com.ubuntu.cloud:released:download": {
		   "datatype": "image-downloads",
		   "path": "streams/v1/com.ubuntu.cloud:released:download.json",
		   "updated": "Wed, 01 May 2013 13:30:37 +0000",
		   "products": [
			"com.ubuntu.cloud:server:12.10:amd64",
			"com.ubuntu.cloud:server:13.04:amd64"
		   ],
		   "format": "products:1.0"
		  },
		  "com.ubuntu.juju:tools": {
		   "updated": "Mon, 05 Aug 2013 11:07:04 +0000",
		   "clouds": [
		    {
			  "region": "us-east-1",
		 	  "endpoint": "https://ec2.us-east-1.amazonaws.com"
		    }
		   ],
		   "cloudname": "aws",
		   "datatype": "juju-tools",
		   "format": "products:1.0",
		   "products": [
		     "com.ubuntu.juju:1.13.0:amd64",
		     "com.ubuntu.juju:1.11.4:arm"
		   ],
		   "path": "streams/v1/tools_metadata.json"
		  }
		 },
		 "updated": "Wed, 01 May 2013 13:31:26 +0000",
		 "format": "index:1.0"
		}
`,
	"/streams/v1/tools_metadata.json": `
{
 "content_id": "com.ubuntu.juju:tools",
 "datatype": "juju-tools",
 "updated": "Tue, 04 Jun 2013 13:50:31 +0000",
 "format": "products:1.0",
 "products": {
  "com.ubuntu.juju:1.13.0:amd64": {
   "version": "1.13.0",
   "arch": "amd64",
   "versions": {
    "20130806": {
     "items": {
      "1130preciseamd64": {
       "release": "precise",
       "size": 2973595,
       "path": "tools/releases/20130806/juju-1.13.0-precise-amd64.tgz",
       "ftype": "tar.gz",
       "sha256": "447aeb6a934a5eaec4f703eda4ef2dde"
      },
      "1130raringamd64": {
       "release": "raring",
       "size": 2973173,
       "path": "tools/releases/20130806/juju-1.13.0-raring-amd64.tgz",
       "ftype": "tar.gz",
       "sha256": "df07ac5e1fb4232d4e9aa2effa57918a"
      }
     }
    }
   }
  },
  "com.ubuntu.juju:1.11.4:arm": {
   "version": "1.11.4",
   "arch": "arm",
   "versions": {
    "20130806": {
     "items": {
      "1114preciseamd64": {
       "release": "precise",
       "size": 1951096,
       "path": "tools/releases/20130806/juju-1.11.4-precise-arm.tgz",
       "ftype": "tar.gz",
       "sha256": "f65a92b3b41311bdf398663ee1c5cd0c"
      },
      "1114raringamd64": {
       "release": "raring",
       "size": 1950327,
       "path": "tools/releases/20130806/juju-1.11.4-raring-arm.tgz",
       "ftype": "tar.gz",
       "sha256": "6472014e3255e3fe7fbd3550ef3f0a11"
      }
     }
    }
   }
  }
 }
}
`,
	"/streams/v1/image_metadata.json": `
{
 "updated": "Wed, 01 May 2013 13:31:26 +0000",
 "content_id": "com.ubuntu.cloud:released:aws",
 "products": {
  "com.ubuntu.cloud:server:12.04:amd64": {
   "release": "precise",
   "version": "12.04",
   "arch": "amd64",
   "region": "au-east-1",
   "endpoint": "https://somewhere",
   "versions": {
    "20121218": {
     "region": "au-east-2",
     "endpoint": "https://somewhere-else",
     "items": {
      "usww1pe": {
       "root_store": "ebs",
       "virt": "pv",
       "id": "ami-26745463"
      },
      "usww2he": {
       "root_store": "ebs",
       "virt": "hvm",
       "id": "ami-442ea674",
       "region": "us-east-1",
       "endpoint": "https://ec2.us-east-1.amazonaws.com"
      },
      "usww3he": {
       "root_store": "ebs",
       "virt": "hvm",
       "crsn": "uswest3",
       "id": "ami-442ea675"
      }
     },
     "pubname": "ubuntu-precise-12.04-amd64-server-20121218",
     "label": "release"
    },
    "20111111": {
     "items": {
      "usww3pe": {
       "root_store": "ebs",
       "virt": "pv",
       "id": "ami-26745464"
      },
      "usww2pe": {
       "root_store": "instance",
       "virt": "pv",
       "id": "ami-442ea684",
       "region": "us-east-1",
       "endpoint": "https://ec2.us-east-1.amazonaws.com"
      }
     },
     "pubname": "ubuntu-precise-12.04-amd64-server-20111111",
     "label": "release"
    }
   }
  },
  "com.ubuntu.cloud:server:12.04:arm": {
   "release": "precise",
   "version": "12.04",
   "arch": "arm",
   "region": "us-east-1",
   "endpoint": "https://ec2.us-east-1.amazonaws.com",
   "versions": {
    "20121219": {
     "items": {
      "usww2he": {
       "root_store": "ebs",
       "virt": "pv",
       "id": "ami-442ea699"
      }
     },
     "pubname": "ubuntu-precise-12.04-arm-server-20121219",
     "label": "release"
    }
   }
  }
 },
 "_aliases": {
  "crsn": {
   "uswest3": {
    "region": "us-west-3",
    "endpoint": "https://ec2.us-west-3.amazonaws.com"
   }
  }
 },
 "format": "products:1.0"
}
`,
}

type TestDataSuite struct {
	testRoundTripper *jujutest.ProxyRoundTripper
}

func (s *TestDataSuite) SetUpSuite(c *gc.C) {
	s.testRoundTripper = &jujutest.ProxyRoundTripper{}
	s.testRoundTripper.RegisterForScheme("test")
	s.testRoundTripper.Sub = jujutest.NewCannedRoundTripper(
		imageData, map[string]int{"test://unauth": http.StatusUnauthorized})
}

func (s *TestDataSuite) TearDownSuite(c *gc.C) {
	s.testRoundTripper.Sub = nil
}

type LocalLiveSimplestreamsSuite struct {
	testing.LoggingSuite
	BaseURL         string
	RequireSigned   bool
	DataType        string
	ValidConstraint simplestreams.LookupConstraint
}

func (s *LocalLiveSimplestreamsSuite) SetUpSuite(c *gc.C) {
	s.LoggingSuite.SetUpSuite(c)
}

func (s *LocalLiveSimplestreamsSuite) TearDownSuite(c *gc.C) {
	s.LoggingSuite.TearDownSuite(c)
}

const (
	Index_v1   = "index:1.0"
	Product_v1 = "products:1.0"
)

type testConstraint struct {
	simplestreams.LookupParams
}

func NewTestConstraint(params simplestreams.LookupParams) *testConstraint {
	return &testConstraint{LookupParams: params}
}

func (tc *testConstraint) Ids() ([]string, error) {
	version, err := simplestreams.SeriesVersion(tc.Series)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(tc.Arches))
	for i, arch := range tc.Arches {
		ids[i] = fmt.Sprintf("com.ubuntu.cloud:server:%s:%s", version, arch)
	}
	return ids, nil
}

func init() {
	// Ensure out test struct can have its tags extracted.
	simplestreams.RegisterStructTags(TestItem{})
}

type TestItem struct {
	Id          string `json:"id"`
	Storage     string `json:"root_store"`
	VType       string `json:"virt"`
	Arch        string `json:"arch"`
	RegionAlias string `json:"crsn"`
	RegionName  string `json:"region"`
	Endpoint    string `json:"endpoint"`
}

func (s *LocalLiveSimplestreamsSuite) indexPath() string {
	if s.RequireSigned {
		return simplestreams.DefaultIndexPath + simplestreams.SignedSuffix
	}
	return simplestreams.DefaultIndexPath + simplestreams.UnsignedSuffix
}

func (s *LocalLiveSimplestreamsSuite) TestGetIndex(c *gc.C) {
	indexRef, err := s.GetIndexRef(Index_v1)
	c.Assert(err, gc.IsNil)
	c.Assert(indexRef.Format, gc.Equals, Index_v1)
	c.Assert(indexRef.BaseURL, gc.Equals, s.BaseURL)
	c.Assert(len(indexRef.Indexes) > 0, gc.Equals, true)
}

func (s *LocalLiveSimplestreamsSuite) GetIndexRef(format string) (*simplestreams.IndexReference, error) {
	params := simplestreams.ValueParams{
		DataType:      s.DataType,
		ValueTemplate: TestItem{},
	}
	return simplestreams.GetIndexWithFormat(s.BaseURL, s.indexPath(), format, s.RequireSigned, params)
}

func (s *LocalLiveSimplestreamsSuite) TestGetIndexWrongFormat(c *gc.C) {
	_, err := s.GetIndexRef("bad")
	c.Assert(err, gc.NotNil)
}

func (s *LocalLiveSimplestreamsSuite) TestGetProductsPathExists(c *gc.C) {
	indexRef, err := s.GetIndexRef(Index_v1)
	c.Assert(err, gc.IsNil)
	path, err := indexRef.GetProductsPath(s.ValidConstraint)
	c.Assert(err, gc.IsNil)
	c.Assert(path, gc.Not(gc.Equals), "")
}

func (s *LocalLiveSimplestreamsSuite) TestGetProductsPathInvalidCloudSpec(c *gc.C) {
	indexRef, err := s.GetIndexRef(Index_v1)
	c.Assert(err, gc.IsNil)
	ic := NewTestConstraint(simplestreams.LookupParams{
		CloudSpec: simplestreams.CloudSpec{"bad", "spec"},
	})
	_, err = indexRef.GetProductsPath(ic)
	c.Assert(err, gc.NotNil)
}

func (s *LocalLiveSimplestreamsSuite) TestGetProductsPathInvalidProductSpec(c *gc.C) {
	indexRef, err := s.GetIndexRef(Index_v1)
	c.Assert(err, gc.IsNil)
	ic := NewTestConstraint(simplestreams.LookupParams{
		CloudSpec: s.ValidConstraint.Params().CloudSpec,
		Series:    "precise",
		Arches:    []string{"bad"},
		Stream:    "spec",
	})
	_, err = indexRef.GetProductsPath(ic)
	c.Assert(err, gc.NotNil)
}

func (s *LocalLiveSimplestreamsSuite) AssertGetMetadata(c *gc.C) *simplestreams.CloudMetadata {
	indexRef, err := s.GetIndexRef(Index_v1)
	c.Assert(err, gc.IsNil)
	metadata, err := indexRef.GetCloudMetadataWithFormat(s.ValidConstraint, Product_v1, s.RequireSigned)
	c.Assert(err, gc.IsNil)
	c.Assert(metadata.Format, gc.Equals, Product_v1)
	c.Assert(len(metadata.Products) > 0, gc.Equals, true)
	return metadata
}

func (s *LocalLiveSimplestreamsSuite) TestGetCloudMetadataWithFormat(c *gc.C) {
	s.AssertGetMetadata(c)
}

func (s *LocalLiveSimplestreamsSuite) AssertGetItemCollections(c *gc.C, version string) *simplestreams.ItemCollection {
	metadata := s.AssertGetMetadata(c)
	metadataCatalog := metadata.Products["com.ubuntu.cloud:server:12.04:amd64"]
	ic := metadataCatalog.Items[version]
	return ic
}