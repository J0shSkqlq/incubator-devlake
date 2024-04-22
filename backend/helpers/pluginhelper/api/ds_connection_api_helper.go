/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"strconv"

	"github.com/apache/incubator-devlake/server/api/shared"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/srvhelper"
)

// DsAnyConnectionApiHelper
type DsAnyConnectionApiHelper struct {
	*AnyModelApiHelper
	*srvhelper.AnyConnectionSrvHelper
}

func NewAnyDsConnectionApiHelper(
	basicRes context.BasicRes,
	connSrvHelper *srvhelper.AnyConnectionSrvHelper,
	sterilizer func(c any) any,
) *DsAnyConnectionApiHelper {
	return &DsAnyConnectionApiHelper{
		AnyModelApiHelper:      NewAnyModelApiHelper(basicRes, connSrvHelper.AnyModelSrvHelper, []string{"connectionId"}, sterilizer),
		AnyConnectionSrvHelper: connSrvHelper,
	}
}

func (connApi *DsAnyConnectionApiHelper) GetMergedConnectionAny(input *plugin.ApiResourceInput) (any, errors.Error) {
	connection, err := connApi.FindByPkAny(input)
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "find connection from db")
	}
	if input.Body != nil {
		if err := DecodeMapStruct(input.Body, connection, false); err != nil {
			return nil, err
		}
	}
	return connection, nil
}

func (connApi *DsAnyConnectionApiHelper) Delete(input *plugin.ApiResourceInput) (out *plugin.ApiResourceOutput, err errors.Error) {
	conn, err := connApi.FindByPkAny(input)
	if err != nil {
		return nil, err
	}
	refs, err := connApi.DeleteConnectionAny(conn)
	if err != nil {
		return &plugin.ApiResourceOutput{Body: &shared.ApiBody{
			Success: false,
			Message: err.Error(),
			Data:    refs,
		}, Status: err.GetType().GetHttpCode()}, err
	}
	conn = connApi.Sanitize(conn)
	return &plugin.ApiResourceOutput{
		Body: conn,
	}, nil
}

func extractConnectionId(input *plugin.ApiResourceInput) (uint64, errors.Error) {
	connectionId, ok := input.Params["connectionId"]
	if !ok {
		return 0, errors.BadInput.New("connectionId is required")
	}
	id, err := strconv.ParseUint(connectionId, 10, 64)
	if err != nil {
		return 0, errors.BadInput.Wrap(err, "connectionId must be a number")
	}
	return id, nil
}

type DsConnectionApiHelper[C dal.Tabler] struct {
	*DsAnyConnectionApiHelper
	*ModelApiHelper[C]
}

func NewDsConnectionApiHelper[C dal.Tabler](
	anyConnectionApiHelper *DsAnyConnectionApiHelper,
) *DsConnectionApiHelper[C] {
	return &DsConnectionApiHelper[C]{
		DsAnyConnectionApiHelper: anyConnectionApiHelper,
		ModelApiHelper:           NewModelApiHelper[C](anyConnectionApiHelper.AnyModelApiHelper),
	}
}

func (connApi *DsConnectionApiHelper[C]) GetMergedConnection(input *plugin.ApiResourceInput) (*C, errors.Error) {
	connection, err := connApi.GetMergedConnectionAny(input)
	return connection.(*C), err
}
