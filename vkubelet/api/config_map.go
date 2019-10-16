/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package api

import (
	"bytes"
	"encoding/gob"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sync"
)

type KeyValueStore struct {
	configMapInterface corev1.ConfigMapInterface
	mutex              *sync.Mutex
	name               string
	configMap          *v1.ConfigMap
}

func NewKeyValueStore(configMapInterface corev1.ConfigMapInterface, storeName string) (*KeyValueStore, error) {
	store := &KeyValueStore{
		configMapInterface: configMapInterface,
		mutex:              &sync.Mutex{},
		name:               storeName,
		configMap:          nil,
	}

	if configMap, err := store.getStore(); err != nil {
		return nil, err
	} else {
		store.configMap = configMap
	}

	if store.configMap.BinaryData == nil {
		store.configMap.BinaryData = make(map[string][]byte)
	}

	return store, nil
}

func (store *KeyValueStore) Get(key string, target interface{}) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if data := store.configMap.BinaryData[key]; data == nil {
		return nil
	} else {
		return store.decode(data, target)
	}
}

func (store *KeyValueStore) Put(key string, value interface{}) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	data, err := store.encode(value)
	if err != nil {
		return err
	}

	store.configMap.BinaryData[key] = data
	_, err = store.configMapInterface.Update(store.configMap)

	return err
}

func (store *KeyValueStore) Remove(key string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	delete(store.configMap.BinaryData, key)
	_, err := store.configMapInterface.Update(store.configMap)

	return err
}

func (store *KeyValueStore) Keys() []string {
	keys := make([]string, 0, len(store.configMap.BinaryData))
	for key := range store.configMap.BinaryData {
		keys = append(keys, key)
	}

	return keys
}

func (store *KeyValueStore) Size() int {
	return len(store.configMap.BinaryData)
}

func (store *KeyValueStore) getStore() (*v1.ConfigMap, error) {
	var configMap *v1.ConfigMap
	var err error

	configMap, err = store.configMapInterface.Get(store.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			configMap, err = store.createStore()
		}

		if err != nil {
			return nil, err
		}
	}

	return configMap, nil
}

func (store *KeyValueStore) createStore() (*v1.ConfigMap, error) {
	cfgMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: store.name,
		},
		Data: map[string]string{},
	}

	return store.configMapInterface.Create(cfgMap)
}

func (store *KeyValueStore) encode(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (store *KeyValueStore) decode(data []byte, target interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(target)
}
