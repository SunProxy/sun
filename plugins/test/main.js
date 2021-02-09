//checked after loading to decide wether we should change the handler...
events = true

function packet_receive(ctx, pk, ray) {
    logger.Infof("Received packet id: %T", pk)
}

function packet_send(ctx, pk, ray) {
    logger.Infof("Sent packet id: %T", pk)
}