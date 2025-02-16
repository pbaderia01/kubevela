/*
Copyright 2022 The KubeVela Authors.

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

package application

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	"github.com/oam-dev/kubevela/pkg/auth"
	"github.com/oam-dev/kubevela/pkg/features"
	"github.com/oam-dev/kubevela/pkg/oam"
)

// MutatingHandler adding user info to application annotations
type MutatingHandler struct {
	Decoder *admission.Decoder
}

var _ admission.Handler = &MutatingHandler{}

// Handle mutate application
func (h *MutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if !utilfeature.DefaultMutableFeatureGate.Enabled(features.AuthenticateApplication) {
		return admission.Patched("")
	}

	if slices.Contains(req.UserInfo.Groups, common.Group) {
		return admission.Patched("")
	}

	app := &v1beta1.Application{}
	if err := h.Decoder.Decode(req, app); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if metav1.HasAnnotation(app.ObjectMeta, oam.AnnotationApplicationServiceAccountName) {
		return admission.Errored(http.StatusBadRequest, errors.New("service-account annotation is not permitted when authentication enabled"))
	}
	auth.SetUserInfoInAnnotation(&app.ObjectMeta, req.UserInfo)

	bs, err := json.Marshal(app)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, bs)
}

var _ admission.DecoderInjector = &MutatingHandler{}

// InjectDecoder .
func (h *MutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}

// RegisterMutatingHandler will register component mutation handler to the webhook
func RegisterMutatingHandler(mgr manager.Manager) {
	server := mgr.GetWebhookServer()
	server.Register("/mutating-core-oam-dev-v1beta1-applications", &webhook.Admission{Handler: &MutatingHandler{}})
}
