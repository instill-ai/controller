import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

let proto
let tHost, mgHost, pHost, cHost, mHost, ctHost
let tPublicPort, mgPublicPort, mgPrivatePort, pPublicPort, pPrivatePort, cPublicPort, cPrivatePort, mPublicPort, mPrivatePort, ctPrivatePort

if (__ENV.API_GATEWAY_HOST && !__ENV.API_GATEWAY_PORT || !__ENV.API_GATEWAY_HOST && __ENV.API_GATEWAY_PORT) {
  fail("both API_GATEWAY_HOST and API_GATEWAY_PORT should be properly configured.")
}

export const apiGatewayMode = (__ENV.API_GATEWAY_HOST && __ENV.API_GATEWAY_PORT);

if (__ENV.API_GATEWAY_PROTOCOL) {
  if (__ENV.API_GATEWAY_PROTOCOL !== "http" && __ENV.API_GATEWAY_PROTOCOL != "https") {
    fail("only allow `http` or `https` for API_GATEWAY_PROTOCOL")
  }
  proto = __ENV.API_GATEWAY_PROTOCOL
} else {
  proto = "http"
}
if (apiGatewayMode) {
  // api-gateway mode
  pHost = cHost = mHost = tHost = mgHost = ctHost = __ENV.API_GATEWAY_HOST
  pPrivatePort = 3081
  cPrivatePort = 3082
  mPrivatePort = 3083
  mgPrivatePort = 3084
  ctPrivatePort = 3085
  tPublicPort = mgPublicPort = pPublicPort = cPublicPort = mPublicPort = 8080
} else {
  // direct microservice mode
  tHost = "triton-server"
  mgHost = "mgmt-backend"
  pHost = "pipeline-backend"
  cHost = "connector-backend"
  mHost = "model-backend"
  ctHost = "controller"
  pPrivatePort = 3081
  cPrivatePort = 3082
  mPrivatePort = 3083
  mgPrivatePort = 3084
  ctPrivatePort = 3085
  tPublicPort = 8001
  pPublicPort = 8081
  cPublicPort = 8082
  mPublicPort = 8083
  mgPublicPort = 8084
}

export const connectorPublicHost = `${proto}://${cHost}:${cPublicPort}`;
export const connectorPrivateHost = `${proto}://${cHost}:${cPrivatePort}`;
export const pipelinePublicHost = `${proto}://${pHost}:${pPublicPort}`;
export const pipelinePrivateHost = `${proto}://${pHost}:${pPrivatePort}`;
export const modelPublicHost = `${proto}://${mHost}:${mPublicPort}`;
export const modelPrivateHost = `${proto}://${mHost}:${mPrivatePort}`;
export const mgmtPublicHost = `${proto}://${mgHost}:${mgPublicPort}`;
export const mgmtPrivateHost = `${proto}://${mgHost}:${mgPrivatePort}`;
export const tritonPublicHost = `${proto}://${tHost}:${tPublicPort}`;
export const controllerPrivateHost = `${proto}://${ctHost}:${ctPrivatePort}`;

export const controllerGRPCPrivateHost = `${ctHost}:${ctPrivatePort}`;

export const modelResourcePermalink = `resources/${uuidv4()}/types/models`
export const modelName = "models/model-name"

export const sourceConnectorResourcePermalink = `resources/${uuidv4()}/types/source-connectors`
export const sourceConnectorName = "source-connectors/source-connector-name"

export const destinationConnectorResourcePermalink = `resources/${uuidv4()}/types/destination-connectors`
export const destinationConnectorName = "destination-connectors/destination-connector-name"

export const pipelineResourcePermalink = `resources/${uuidv4()}/types/pipelines`
export const pipelineName = "pipelines/pipeline-name"

export const serviceResourceName = "resources/model-backend/types/services"
export const serviceName = "model-backend"
