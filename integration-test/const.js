let proto
let tHost, mgHost, pHost, cHost, mHost, ctHost
let tPublicPort, mgPublicPort, mgPrivatePort, pPublicPort, pPrivatePort, cPublicPort, cPrivatePort, mPublicPort, mPrivatePort, ctPrivatePort

if (__ENV.MODE == "api-gateway") {
  // api-gateway mode
  proto = "http"
  pHost = cHost = mHost = tHost = mgHost = ctHost = "api-gateway"
  pPrivatePort = 3081
  cPrivatePort = 3082
  mPrivatePort = 3083
  mgPrivatePort = 3084
  ctPrivatePort = 3085
  tPublicPort = mgPublicPort = pPublicPort = cPublicPort = mPublicPort = 8080
} else if (__ENV.MODE == "localhost") {
  // localhost mode for GitHub Actions
  proto = "http"
  pHost = cHost = mHost = tHost = mgHost = ctHost = "localhost"
  pPrivatePort = 3081
  cPrivatePort = 3082
  mPrivatePort = 3083
  mgPrivatePort = 3084
  ctPrivatePort = 3085
  tPublicPort = mgPublicPort = pPublicPort = cPublicPort = mPublicPort = 8080
} else {
  // direct microservice mode
  proto = "http"
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

export const modelResourceName = "resources/this-is-model-name/types/models"
export const modelName = "models/model-name"

export const sourceConnectorResourceName = "resources/source-connector-name/types/source-connectors"
export const sourceConnectorName = "source-connectors/source-connector-name"

export const destinationConnectorResourceName = "resources/destination-connector-name/types/destination-connectors"
export const destinationConnectorName = "destination-connectors/destination-connector-name"

export const pipelineResourceName = "resources/pipeline-name/types/pipelines"
export const pipelineName = "pipelines/pipeline-name"

export const serviceResourceName = "resources/model-backend/types/services"
export const serviceName = "model-backend"
