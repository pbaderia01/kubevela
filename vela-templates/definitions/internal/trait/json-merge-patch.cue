"json-merge-patch": {
	type: "trait"
	annotations: {}
	labels: {}
	description: "Patch the output following Json Merge Patch strategy, following RFC 7396."
	attributes: {
		podDisruptive: true
		appliesToWorkloads: ["*"]
	}
}
template: {
	parameter: {...}
	// +patchStrategy=jsonMergePatch
	patch: parameter
}
