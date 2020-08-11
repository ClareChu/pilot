# pilot
将docker-compose to k8s app struts


## 项目初始化

演示
1,初始化项目
项目使用go.mod管理,所以在初始化项目的同时,我们需要初始化依赖库

#创建目录
```bash
$ mkdir github.com/ClareChu/pilot && cd github.com/ClareChu/pilot
```
#初始化项目
```bash
$ go mod init github.com/ClareChu/pilot
```
# 获取依赖
```bash
$ go get k8s.io/apimachinery@v0.0.0-20190425132440-17f84483f500
$ go get k8s.io/client-go@v0.0.0-20190425172711-65184652c889
$ go get k8s.io/code-generator@v0.0.0-20190419212335-ff26e7842f9d
```

2,初始化crd资源类型
在初始化了项目后,需要建立好自己的crd struct,然后使用code-generator生成我们的代码.

$ mkdir -p api/samplecontroller/v1alpha1 && cd api/samplecontroller/v1alpha1
此处我们的自定义资源的group为samplecontroller,版本为v1alpha1

在文件夹中新建:

doc.go
```go
// +k8s:deepcopy-gen=package
// +groupName=samplecontroller.k8s.io

// v1alpha1版本的api包

package v1alpha1
```

types.go

```go

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo is a specification for a Foo resource
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FooSpec   `json:"spec"`
	Status FooStatus `json:"status"`
}

// FooSpec is the spec for a Foo resource
type FooSpec struct {
	DeploymentName string `json:"deploymentName"`
	Replicas       *int32 `json:"replicas"`
}

// FooStatus is the status for a Foo resource
type FooStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooList is a list of Foo resources
type FooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Foo `json:"items"`
}
```

3,创建生成脚本
有了crd的定义后,我们需要准备我们的构建脚本和对依赖进行一定的修改.

```bash
$ mkdir hack && cd hack
```

建立tools.go来依赖code-generator,因为在没有代码使用code-generator时,go module 默认不会为我们依赖此包.

```bash


// +build tools

/*
Copyright 2019 The Kubernetes Authors.

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

// This package imports things required by build scripts, to force `go mod` to see them as dependencies
package tools

import _ "k8s.io/code-generator"
```

同时 编写我们的构建脚本:

update-codegen.sh

```bash

#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
../vendor/k8s.io/code-generator/generate-groups.sh \
  "deepcopy,client,informer,lister" \
  github.com/ClareChu/pilot/generated \
  github.com/ClareChu/pilot/api \
  samplecontroller:v1alpha1 \
  --go-header-file $(pwd)/boilerplate.go.txt \
  --output-base $(pwd)/../../../../


可以看到generate-groups.sh其中有几个参数,使用命令可以看到如下:

Usage: generate-groups.sh <generators> <output-package> <apis-package> <groups-versions> ...

  <generators>        the generators comma separated to run (deepcopy,defaulter,client,lister,informer) or "all".
  <output-package>    the output package name (e.g. github.com/example/project/pkg/generated).
  <apis-package>      the external types dir (e.g. github.com/example/api or github.com/example/project/pkg/apis).
  <groups-versions>   the groups and their versions in the format "groupA:v1,v2 groupB:v1 groupC:v2", relative
                      to <api-package>.
  ...                 arbitrary flags passed to all generator binaries.


Examples:
  generate-groups.sh all             github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"
  generate-groups.sh deepcopy,client github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"

```
在构建api时,我们还提供了文件头,所以我们在此也创建文件头:

boilerplate.go.txt


```text
/*
Copyright The Kubernetes Authors.

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
当然,这个文件的头是可以定制的
```

4,生成api
当我们做好这些准备工作后就可以开始生成我们的crd资源的clientset等api了.

```bash
# 生成vendor文件夹
$ go mod vendor
# 进入项目根目录,为vendor中的code-generator赋予权限
$ chmod -R 777 vendor
# 调用脚本生成代码
$ cd hack && ./update-codegen.sh
Generating deepcopy funcs
Generating clientset for samplecontroller:v1alpha1 at code-generator-test/generated/clientset
Generating listers for samplecontroller:v1alpha1 at code-generator-test/generated/listers
Generating informers for samplecontroller:v1alpha1 at code-generator-test/generated/informers

```

仔细观察,发现code-generator-test/api/samplecontroller/v1alpha1下多出了一个zz_generated.deepcopy.go的文件,在generated文件夹下生成了clientset和informers和listers三个文件夹

5,使用
在生成了客户端代码后,我们还是需要手动的注册这个crd资源,才能正真使用这个client,不然在编译时会出现如下错误
```bash
# code-generator-test/generated/clientset/versioned/scheme
generated/clientset/versioned/scheme/register.go:35:2: undefined: v1alpha1.AddToScheme
# code-generator-test/generated/listers/samplecontroller/v1alpha1
generated/listers/samplecontroller/v1alpha1/foo.go:92:34: undefined: v1alpha1.Resource
```
由编译的错误提示,可以看到,需要提供v1alpha1.AddToScheme和v1alpha1.Resource这两个变量供client注册.

所以我们还需要在v1alpha1下新建一个register.go文件,内容如下:

```go

/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
// 注册自己的自定义资源
var SchemeGroupVersion = schema.GroupVersion{Group: "samplecontroller.k8s.io", Version: "v1alpha1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	//注意,添加了foo/foolist 两个资源到scheme
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Foo{},
		&FooList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
```

