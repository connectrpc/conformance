// Copyright 2022 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import TestCases from "./test-cases";

const root = document.getElementById("root")
if (root === null) {
    throw "root is not existed"
}

// TODO(doria): figure out how to pass the host and port through config.
ReactDOM.createRoot(root).render(
  <React.StrictMode>
    <TestCases host="localhost" port="9092" />
  </React.StrictMode>
);
