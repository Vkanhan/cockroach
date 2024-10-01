// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

interface ObjectConstructor {
  assign(target: any, ...sources: any[]): any;
}
