import grpc from 'k6/net/grpc';
import {
    check,
    group
} from "k6";

import * as constant from "./const.js"

const clientPrivate = new grpc.Client();
clientPrivate.load(['proto/vdp/controller/v1alpha'], 'controller_service.proto');

export function CheckModelResource() {
    var httpModelResource = {
        "resource_permalink": constant.modelResourcePermalink,
        "model_state": "STATE_ONLINE"
    }

    clientPrivate.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
    });

    group("Controller API: Create model resource state in etcd", () => {
        var resCreateModelHTTP = clientPrivate.invoke('vdp.controller.v1alpha.ControllerPrivateService/UpdateResource', {
            resource: httpModelResource
        })

        console.log(resCreateModelHTTP)

        console.log(httpModelResource.resource_permalink)

        check(resCreateModelHTTP, {
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response StatusOK": (r) => r.status === grpc.StatusOK,
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response modelResource resource_permalink matched": (r) => r.message.resource.resourcePermalink == httpModelResource.resource_permalink,
        });
    });

    group("Controller API: Get model resource state in etcd", () => {
        var resGetModelHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            resource_permalink: httpModelResource.resource_permalink
        })

        check(resGetModelHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpModelResource.resource_permalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpModelResource.resource_permalink} response modelResource resource_permalink matched`]: (r) => r.message.resource.resourcePermalink === httpModelResource.resource_permalink,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpModelResource.resource_permalink} response modelResource state matched STATE_ONLINE`]: (r) => r.message.resource.modelState == "STATE_ONLINE",
        });
    });
}

export function CheckSourceConnectorResource() {
    var httpSourceConnectorResource = {
        "resource_permalink": constant.sourceConnectorResourcePermalink,
        "connector_state": "STATE_CONNECTED"
    }

    clientPrivate.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
    });

    group("Controller API: Create source connector resource state in etcd", () => {
        var resCreateSourceConnectorHTTP = clientPrivate.invoke('vdp.controller.v1alpha.ControllerPrivateService/UpdateResource', {
            resource: httpSourceConnectorResource
        })

        check(resCreateSourceConnectorHTTP, {
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response StatusOK": (r) => r.status === grpc.StatusOK,
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response connectorResource resource_permalink matched": (r) => r.message.resource.resourcePermalink == httpSourceConnectorResource.resource_permalink,
        });
    });

    group("Controller API: Get source connector resource state in etcd", () => {
        var resGetSourceConnectorHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            resource_permalink: httpSourceConnectorResource.resource_permalink
        })

        check(resGetSourceConnectorHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpSourceConnectorResource.resource_permalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpSourceConnectorResource.resource_permalink} response connectorResource resource_permalink matched`]: (r) => r.message.resource.resourcePermalink === httpSourceConnectorResource.resource_permalink,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpSourceConnectorResource.resource_permalink} response connectorResource state matched STATE_CONNECTED`]: (r) => r.message.resource.connectorState == "STATE_CONNECTED",
        });
    });
}

export function CheckDestinationConnectorResource() {
    var httpDestinationConnectorResource = {
        "resource_permalink": constant.destinationConnectorResourcePermalink,
        "connector_state": "STATE_CONNECTED"
    }

    clientPrivate.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
    });

    group("Controller API: Create destination connector resource state in etcd", () => {
        var resCreatpDestinationConnectorHTTP = clientPrivate.invoke('vdp.controller.v1alpha.ControllerPrivateService/UpdateResource', {
            resource: httpDestinationConnectorResource
        })

        check(resCreatpDestinationConnectorHTTP, {
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response StatusOK": (r) => r.status === grpc.StatusOK,
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response connectorResource resource_permalink matched": (r) => r.message.resource.resourcePermalink == httpDestinationConnectorResource.resource_permalink,
        });
    });

    group("Controller API: Get destination connector resource state in etcd", () => {
        var resGetDestinationConnectorHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            resource_permalink: httpDestinationConnectorResource.resource_permalink
        })

        check(resGetDestinationConnectorHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpDestinationConnectorResource.resource_permalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpDestinationConnectorResource.resource_permalink} response connectorResource resource_permalink matched`]: (r) => r.message.resource.resourcePermalink === httpDestinationConnectorResource.resource_permalink,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpDestinationConnectorResource.resource_permalink} response connectorResource state matched STATE_CONNECTED`]: (r) => r.message.resource.connectorState == "STATE_CONNECTED",
        });
    });
}

export function CheckPipelineResource() {
    var httpPipelineResource = {
        "resource_permalink": constant.pipelineResourcePermalink,
        "pipeline_state": "STATE_ACTIVE"
    }

    clientPrivate.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
    });

    group("Controller API: Create pipeline resource state in etcd", () => {
        var resCreatePipelineHTTP = clientPrivate.invoke('vdp.controller.v1alpha.ControllerPrivateService/UpdateResource', {
            resource: httpPipelineResource
        })

        check(resCreatePipelineHTTP, {
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response StatusOK": (r) => r.status === grpc.StatusOK,
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response pipeline resource_permalink matched": (r) => r.message.resource.resourcePermalink == httpPipelineResource.resource_permalink,
        });
    });

    group("Controller API: Get pipeline resource state in etcd", () => {
        var resGetPipelineHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            resource_permalink: httpPipelineResource.resource_permalink
        })

        check(resGetPipelineHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpPipelineResource.resource_permalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpPipelineResource.resource_permalink} response pipeline resource_permalink matched`]: (r) => r.message.resource.resourcePermalink === httpPipelineResource.resource_permalink,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpPipelineResource.resource_permalink} response pipeline state matched STATE_ACTIVE`]: (r) => r.message.resource.pipelineState == "STATE_ACTIVE",
        });
    });
}

export function CheckServiceResource() {
    var httpServiceResource = {
        "resource_permalink": constant.serviceResourceName,
        "backend_state": "SERVING_STATUS_SERVING"
    }

    clientPrivate.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
    });

    group("Controller API: Create service resource state in etcd", () => {
        var resCreateServiceHTTP = clientPrivate.invoke('vdp.controller.v1alpha.ControllerPrivateService/UpdateResource', {
            resource: httpServiceResource
        })

        check(resCreateServiceHTTP, {
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response StatusOK": (r) => r.status === grpc.StatusOK,
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response service name matched": (r) => r.message.resource.name == httpServiceResource.name,
        });
    });

    group("Controller API: Get service resource state in etcd", () => {
        var resGetServiceHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            resource_permalink: httpServiceResource.resource_permalink
        })

        check(resGetServiceHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpServiceResource.resource_permalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpServiceResource.resource_permalink} response service name matched`]: (r) => r.message.resource.resourcePermalink === httpServiceResource.resource_permalink,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpServiceResource.resource_permalink} response service state matched STATE_ACTIVE`]: (r) => r.message.resource.backendState == "SERVING_STATUS_SERVING",
        });
    });
}
