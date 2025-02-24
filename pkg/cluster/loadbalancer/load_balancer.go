/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package loadbalancer

import (
	"github.com/apache/dubbo-go-pixiu/pkg/model"
)

type LoadBalancer interface {
	Handler(c *model.ClusterConfig, policy model.LbPolicy) *model.Endpoint
}

// LoadBalancerStrategy load balancer strategy mode
var LoadBalancerStrategy = map[model.LbPolicyType]LoadBalancer{}

func RegisterLoadBalancer(name model.LbPolicyType, balancer LoadBalancer) {
	if _, ok := LoadBalancerStrategy[name]; ok {
		panic("load balancer register fail " + name)
	}
	LoadBalancerStrategy[name] = balancer
}

func RegisterConsistentHashInit(name model.LbPolicyType, function model.ConsistentHashInitFunc) {
	if _, ok := model.ConsistentHashInitMap[name]; ok {
		panic("consistent hash load balancer register fail " + name)
	}
	model.ConsistentHashInitMap[name] = function
}
