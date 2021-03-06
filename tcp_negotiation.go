package ravendb

import (
	"encoding/json"
	"io"
	"strconv"
)

const (
	outOfRangeStatus = -1
	dropStatus       = -2
)

func negotiateProtocolVersion(stream io.Writer, parameters *tcpNegotiateParameters) (*supportedFeatures, error) {
	v := parameters.version
	currentRef := &v
	for {
		sendTcpVersionInfo(stream, parameters, *currentRef)
		version := parameters.readResponseAndGetVersionCallback(parameters.destinationUrl)

		/*
			   if (logger.isInfoEnabled()) {
				   logger.info("Read response from " + ObjectUtils.firstNonNull(parameters.getSourceNodeTag(), parameters.getDestinationUrl()) + " for " + parameters.getOperation() + ", received version is '" + version + "'");
			   }
		*/

		if version == *currentRef {
			break
		}

		//In this case we usually throw internally but for completeness we better handle it
		if version == dropStatus {
			return getSupportedFeaturesFor(operationDrop, dropBaseLine), nil
		}

		status := operationVersionSupported(parameters.operation, version, currentRef)

		if status == supportedStatus_OUT_OF_RANGE {
			sendTcpVersionInfo(stream, parameters, outOfRangeStatus)
			return nil, newIllegalArgumentError("The " + parameters.operation + " version " + strconv.Itoa(parameters.version) + " is out of range, out lowest version is " + strconv.Itoa(*currentRef))
		}

		/*
		   if (logger.isInfoEnabled()) {
		       logger.info("The version " +  version + " is " + status + ", will try to agree on '"
		               + currentRef.value + "' for " + parameters.getOperation() + " with "
		               + ObjectUtils.firstNonNull(parameters.getDestinationNodeTag(), parameters.getDestinationUrl()));
		   }
		*/
	}
	/*
		   if (logger.isInfoEnabled()) {
			   logger.info(ObjectUtils.firstNonNull(parameters.getDestinationNodeTag(), parameters.getDestinationUrl()) + " agreed on version " + currentRef.value + " for " + parameters.getOperation());
		   }
	*/
	return getSupportedFeaturesFor(parameters.operation, *currentRef), nil
}

func sendTcpVersionInfo(stream io.Writer, parameters *tcpNegotiateParameters, currentVersion int) error {
	/*
		if (logger.isInfoEnabled()) {
			logger.info("Send negotiation for " + parameters.getOperation() + " in version " + currentVersion);
		}
	*/
	m := map[string]interface{}{
		"DatabaseName":     parameters.database,
		"Operation":        parameters.operation,
		"SourceNodeTag":    parameters.sourceNodeTag,
		"OperationVersion": currentVersion,
	}
	enc := json.NewEncoder(stream)
	return enc.Encode(m)
}
