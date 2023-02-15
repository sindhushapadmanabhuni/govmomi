/*
Copyright (c) 2018-2022 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package library

import (
	"context"
	"flag"
	"fmt"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vapi/library"
)

type create struct {
	*flags.DatastoreFlag
	library library.Library
	sub     library.Subscription
	pub     library.Publication
}

func init() {
	cli.Register("library.create", &create{})
}

func (cmd *create) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatastoreFlag, ctx = flags.NewDatastoreFlag(ctx)
	cmd.DatastoreFlag.Register(ctx, f)

	cmd.sub.AutomaticSyncEnabled = new(bool)
	cmd.sub.OnDemand = new(bool)

	f.Var(flags.NewOptionalString(&cmd.library.Description), "d", "Description of library")
	f.StringVar(&cmd.sub.SubscriptionURL, "sub", "", "Subscribe to library URL")
	f.StringVar(&cmd.sub.UserName, "sub-username", "", "Subscription username")
	f.StringVar(&cmd.sub.Password, "sub-password", "", "Subscription password")
	f.StringVar(&cmd.sub.SslThumbprint, "thumbprint", "", "SHA-1 thumbprint of the host's SSL certificate")
	f.BoolVar(cmd.sub.AutomaticSyncEnabled, "sub-autosync", true, "Automatic synchronization")
	f.BoolVar(cmd.sub.OnDemand, "sub-ondemand", false, "Download content on demand")
	f.Var(flags.NewOptionalBool(&cmd.pub.Published), "pub", "Publish library")
	f.StringVar(&cmd.pub.UserName, "pub-username", "", "Publication username")
	f.StringVar(&cmd.pub.Password, "pub-password", "", "Publication password")
	f.StringVar(&cmd.library.SecurityPolicyID, "policy", "", "Security Policy ID")
}

func (cmd *create) Usage() string {
	return "NAME"
}

func (cmd *create) Description() string {
	return `Create library.

Examples:
  govc library.create library_name
  govc library.create -sub http://server/path/lib.json library_name
  govc library.create -pub library_name`
}

func (cmd *create) Run(ctx context.Context, f *flag.FlagSet) error {
	if f.NArg() != 1 {
		return flag.ErrHelp
	}

	ds, err := cmd.Datastore()
	if err != nil {
		return err
	}

	cmd.library.Name = f.Arg(0)
	cmd.library.Type = "LOCAL"
	cmd.library.Storage = []library.StorageBackings{
		{
			DatastoreID: ds.Reference().Value,
			Type:        "DATASTORE",
		},
	}

	if cmd.sub.SubscriptionURL != "" {
		cmd.library.Subscription = &cmd.sub
		cmd.library.Type = "SUBSCRIBED"
		cmd.sub.AuthenticationMethod = "NONE"
		if cmd.sub.Password != "" {
			cmd.sub.AuthenticationMethod = "BASIC"
		}
	}

	if cmd.pub.Published != nil && *cmd.pub.Published {
		cmd.library.Publication = &cmd.pub
		cmd.pub.AuthenticationMethod = "NONE"
		if cmd.pub.Password != "" {
			cmd.sub.AuthenticationMethod = "BASIC"
		}
	}

	c, err := cmd.RestClient()
	if err != nil {
		return err
	}

	id, err := library.NewManager(c).CreateLibrary(ctx, cmd.library)
	if err != nil {
		return err
	}

	fmt.Println(id)
	return nil
}
