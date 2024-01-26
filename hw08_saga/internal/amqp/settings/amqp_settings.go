package amqp_settings

const ExchangeOrder = "shop_exchange_order"
const QueueOrder = "shop_queue_order"

const ExchangeStatus = "shop_exchange_status"
const QueueStatusOrderService = "shop_queue_status_for_order_service"
const QueueStatusPayService = "shop_queue_status_for_pay_service"
const QueueStatusStockService = "shop_queue_status_for_stock_service"
const QueueStatusDeliveryService = "shop_queue_status_for_delivery_service"

const RoutingKeyPayService = "shop_for_pay_service"
const RoutingKeyStockService = "shop_for_stock_service"
const RoutingKeyDeliveryService = "shop_for_delivery_service"
