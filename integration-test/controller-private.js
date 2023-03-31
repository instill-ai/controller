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
        "name": constant.modelResourceName,
        "model_state": "STATE_ONLINE"
    }

    clientPrivate.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
    });

    group("Controller API: Create model resource state in etcd", () => {
        var resCreateModelHTTP = clientPrivate.invoke('vdp.controller.v1alpha.ControllerPrivateService/UpdateResource', {
            resource: httpModelResource
        })

        check(resCreateModelHTTP, {
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response StatusOK": (r) => r.status === grpc.StatusOK,
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response modelResource name matched": (r) => r.message.resource.name == httpModelResource.name,
        });
    });

    group("Controller API: Get model resource state in etcd", () => {
        var resGetModelHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            name: httpModelResource.name
        })

        check(resGetModelHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpModelResource.name} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpModelResource.name} response modelResource name matched`]: (r) => r.message.resource.name === httpModelResource.name,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpModelResource.name} response modelResource state matched STATE_ONLINE`]: (r) => r.message.resource.modelState == "STATE_ONLINE",
        });
    });
}

export function CheckSourceConnectorResource() {
    var httpSourceConnectorResource = {
        "name": constant.sourceConnectorResourceName,
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
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response connectorResource name matched": (r) => r.message.resource.name == httpSourceConnectorResource.name,
        });
    });

    group("Controller API: Get source connector resource state in etcd", () => {
        var resGetSourceConnectorHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            name: httpSourceConnectorResource.name
        })

        check(resGetSourceConnectorHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpSourceConnectorResource.name} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpSourceConnectorResource.name} response connectorResource name matched`]: (r) => r.message.resource.name === httpSourceConnectorResource.name,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpSourceConnectorResource.name} response connectorResource state matched STATE_CONNECTED`]: (r) => r.message.resource.connectorState == "STATE_CONNECTED",
        });
    });
}

export function CheckDestinationConnectorResource() {
    var httpDestinationConnectorResource = {
        "name": constant.destinationConnectorResourceName,
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
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response connectorResource name matched": (r) => r.message.resource.name == httpDestinationConnectorResource.name,
        });
    });

    group("Controller API: Get destination connector resource state in etcd", () => {
        var resGetDestinationConnectorHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            name: httpDestinationConnectorResource.name
        })

        check(resGetDestinationConnectorHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpDestinationConnectorResource.name} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpDestinationConnectorResource.name} response connectorResource name matched`]: (r) => r.message.resource.name === httpDestinationConnectorResource.name,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpDestinationConnectorResource.name} response connectorResource state matched STATE_CONNECTED`]: (r) => r.message.resource.connectorState == "STATE_CONNECTED",
        });
    });
}

export function CheckPipelineResource() {
    var httpPipelineResource = {
        "name": constant.pipelineResourceName,
        "pipeline_state": "STATE_ACTIVE"
    }

    clientPrivate.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
    });

    group("Controller API: Create pipeline resource state in etcd", () => {
        var resCreatepipelineHTTP = clientPrivate.invoke('vdp.controller.v1alpha.ControllerPrivateService/UpdateResource', {
            resource: httpPipelineResource
        })

        check(resCreatepipelineHTTP, {
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response StatusOK": (r) => r.status === grpc.StatusOK,
            "vdp.controller.v1alpha.ControllerPrivateService/UpdateResource response pipeline name matched": (r) => r.message.resource.name == httpPipelineResource.name,
        });
    });

    group("Controller API: Get pipeline resource state in etcd", () => {
        var resGetPipelineHTTP = clientPrivate.invoke(`vdp.controller.v1alpha.ControllerPrivateService/GetResource`, {
            name: httpPipelineResource.name
        })

        console.log(resGetPipelineHTTP)

        check(resGetPipelineHTTP, {
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpPipelineResource.name} response StatusOK`]: (r) => r.status === grpc.StatusOK,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpPipelineResource.name} response pipeline name matched`]: (r) => r.message.resource.name === httpPipelineResource.name,
            [`vdp.controller.v1alpha.ControllerPrivateService/GetResource ${httpPipelineResource.name} response pipeline state matched STATE_ACTIVE`]: (r) => r.message.resource.pipelineState == "STATE_ACTIVE",
        });
    });
}
