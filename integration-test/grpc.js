import grpc from 'k6/net/grpc';
import {
  check,
  group
} from 'k6';

import * as constant from "./const.js"
import * as controller_service from './controller-private.js';
const client = new grpc.Client();
client.load(['proto/vdp/controller/v1alpha'], 'controller_service.proto');

export let options = {
  setupTimeout: '10s',
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

export default function (data) {

  /*
   * Controller API - API CALLS
   */
  if (!constant.apiGatewayMode) {
    // Health check
    group("Controller API: Health check", () => {
      client.connect(constant.controllerGRPCPrivateHost, {
        plaintext: true
      });

      check(client.invoke('vdp.controller.v1alpha.ControllerPrivateService/Liveness', {}), {
        'Liveness Status is OK': (r) => r && r.status === grpc.StatusOK,
        'Response status is SERVING_STATUS_SERVING': (r) => r && r.message.healthCheckResponse.status === "SERVING_STATUS_SERVING",
      });

      check(client.invoke('vdp.controller.v1alpha.ControllerPrivateService/Readiness', {}), {
        'Readiness Status is OK': (r) => r && r.status === grpc.StatusOK,
        'Response status is SERVING_STATUS_SERVING': (r) => r && r.message.healthCheckResponse.status === "SERVING_STATUS_SERVING",
      });
      client.close();
    });

    controller_service.CheckModelResource()
    controller_service.CheckSourceConnectorResource()
    controller_service.CheckDestinationConnectorResource()
    controller_service.CheckPipelineResource()
    controller_service.CheckServiceResource()
  }
}

export function teardown(data) {
  if (!constant.apiGatewayMode) {
    client.connect(constant.controllerGRPCPrivateHost, {
      plaintext: true
    });
    group("Controller API: Delete all resources created by the test", () => {

      check(client.invoke(`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource`, {
        resource_permalink: constant.modelResourcePermalink
      }), {
        [`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource ${constant.modelResourcePermalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
      });

      check(client.invoke(`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource`, {
        resource_permalink: constant.sourceConnectorResourcePermalink
      }), {
        [`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource ${constant.sourceConnectorResourcePermalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
      });

      check(client.invoke(`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource`, {
        resource_permalink: constant.destinationConnectorResourcePermalink
      }), {
        [`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource ${constant.destinationConnectorResourcePermalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
      });

      check(client.invoke(`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource`, {
        resource_permalink: constant.pipelineResourcePermalink
      }), {
        [`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource ${constant.pipelineResourcePermalink} response StatusOK`]: (r) => r.status === grpc.StatusOK,
      });

      check(client.invoke(`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource`, {
        resource_permalink: constant.serviceResourceName
      }), {
        [`vdp.controller.v1alpha.ControllerPrivateService/DeleteResource ${constant.serviceResourceName} response StatusOK`]: (r) => r.status === grpc.StatusOK,
      });
    });
    client.close();
  } else {
    console.log("No Public APIs")
  }
}
