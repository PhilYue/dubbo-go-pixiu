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

package zookeeper

import (
	"sync"
	"time"
)

import (
	"github.com/dubbogo/go-zookeeper/zk"
)

import (
	"github.com/apache/dubbo-go-pixiu/pkg/adapter/dubboregistry/registry"
	"github.com/apache/dubbo-go-pixiu/pkg/adapter/dubboregistry/remoting/zookeeper"
	"github.com/apache/dubbo-go-pixiu/pkg/logger"
)

var _ registry.Listener = new(applicationServiceListener)

// applicationServiceListener normally monitors the /services/[:application]
type applicationServiceListener struct {
	urls            []interface{}
	servicePath     string

	serviceName     string

	exit chan struct{}
	wg   sync.WaitGroup

	ds *zookeeperDiscovery
}

// pi serviceName : /services/sc1
// newApplicationServiceListener creates a new zk service listener
func newApplicationServiceListener(path string, serviceName string, ds *zookeeperDiscovery) *applicationServiceListener {
	return &applicationServiceListener{
		servicePath:     path, // pi serviceName : /services/sc1
		exit:            make(chan struct{}),
		ds:              ds,
		serviceName: serviceName,
	}
}

func (asl *applicationServiceListener) WatchAndHandle() {
	defer asl.wg.Done()

	var (
		failTimes  int64 = 0
		delayTimer       = time.NewTimer(ConnDelay * time.Duration(failTimes))
	)
	defer delayTimer.Stop()
	for {
		children, e, err := asl.ds.getClient().GetChildrenW(asl.servicePath)
		// pi servicePath : /services/sc1
		//children, e, err := asl.client.GetChildrenW(asl.servicePath) // pi children : [10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c]
		// error handling
		if err != nil {
			failTimes++
			logger.Infof("watching (path{%s}) = error{%v}", asl.servicePath, err)
			if err == zookeeper.ErrNilChildren {
				return
			}
			if err == zookeeper.ErrNilNode {
				logger.Errorf("watching (path{%s}) got errNilNode,so exit listen", asl.servicePath)
				return
			}
			if failTimes > MaxFailTimes {
				logger.Errorf("Error happens on (path{%s}) exceed max fail times: %v,so exit listen", asl.servicePath, MaxFailTimes)
				return
			}
			delayTimer.Reset(ConnDelay * time.Duration(failTimes))
			<-delayTimer.C
			continue
		}
		failTimes = 0
		if continueLoop := asl.waitEventAndHandlePeriod(children, e); !continueLoop {
			return
		}

	}
}

func (asl *applicationServiceListener) waitEventAndHandlePeriod(children []string, e <-chan zk.Event) bool { // pi children : [10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c]
	tickerTTL := defaultTTL
	ticker := time.NewTicker(tickerTTL)
	defer ticker.Stop()
	asl.handleEvent(children) // pi children : [10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c]
	for {
		select {
		case <-ticker.C:
			//asl.handleEvent(children)
		case zkEvent := <-e:
			logger.Warnf("get a zookeeper e{type:%s, server:%s, path:%s, state:%d-%s, err:%s}",
				zkEvent.Type.String(), zkEvent.Server, zkEvent.Path, zkEvent.State, zookeeper.StateToString(zkEvent.State), zkEvent.Err)
			if zkEvent.Type != zk.EventNodeChildrenChanged {
				return true
			}
			asl.handleEvent(children)
			return true
		case <-asl.exit:
			logger.Warnf("listen(path{%s}) goroutine exit now...", asl.servicePath)
			return false
		}
	}
}

func (asl *applicationServiceListener) handleEvent(children []string) { // pi children : [10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c]

	fetchChildren, err := asl.ds.getClient().GetChildren(asl.servicePath)
	// pi children : [/services/sc1, /services/sc2, /services/sc3]
	//fetchChildren, err := asl.client.GetChildren(asl.servicePath) // pi children : [10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c, 10c59770-c3b3-496b-9845-3ee95fe8e62c]
	if err != nil {
		logger.Warnf("Error when retrieving newChildren in path: %s, Error:%s", asl.servicePath, err.Error())
		return
	}
	discovery := asl.ds
	instanceMap := discovery.instanceMap
	for _, id := range fetchChildren {
		serviceInstance, err := discovery.queryForInstance(asl.serviceName, id)
		if err != nil {
			if err == zk.ErrNoNode {
				discovery.delServiceInstance(serviceInstance)
			}
			logger.Errorf("fail %v", err) // pi retry
			continue
		}

		instance := instanceMap[id]
		if instance != nil {
			// pi update
			//discovery.updateServiceInstance(serviceInstance)

		} else {
			// pi add
			discovery.addServiceInstance(serviceInstance)
		}
	}

	if delInstanceIds := Diff(children, fetchChildren); delInstanceIds != nil {
		for _, id := range delInstanceIds {
			discovery.delServiceInstance(instanceMap[id])
		}
	}

}

// Close closes this listener
func (asl *applicationServiceListener) Close() {
	close(asl.exit)
	asl.wg.Wait()
}
