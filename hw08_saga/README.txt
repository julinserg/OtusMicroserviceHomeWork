1. Манифесты - https://github.com/julinserg/otus-microservice-hw/tree/main/hw08_saga/deployments/kubernetes
2. Перед применением манифестов необходимо убедиться что в minikube включен nginx ingress controller, если нет то включить - minikube addons enable ingress
2. Перед применением манифестов необходимо установить rabbitmq и postgresql следующими командами:
sudo helm install mq-shop oci://registry-1.docker.io/bitnamicharts/rabbitmq --set auth.username='guest',auth.password='guest'
sudo helm install pg-order-service oci://registry-1.docker.io/bitnamicharts/postgresql --set auth.username='postgres',auth.password='postgres',auth.database='shop_order'
sudo helm install pg-pay-service oci://registry-1.docker.io/bitnamicharts/postgresql --set auth.username='postgres',auth.password='postgres',auth.database='shop_pay'
sudo helm install pg-stock-service oci://registry-1.docker.io/bitnamicharts/postgresql --set auth.username='postgres',auth.password='postgres',auth.database='shop_stock'
sudo helm install pg-delivery-service oci://registry-1.docker.io/bitnamicharts/postgresql --set auth.username='postgres',auth.password='postgres',auth.database='shop_delivery'
3. Коллекция тестов постман - https://github.com/julinserg/otus-microservice-hw/blob/main/hw08_saga/test/postman/SagaTest.json
4. Исходный код проекта - https://github.com/julinserg/otus-microservice-hw/tree/main/hw08_saga

Описание реализации:

1. Для реализации распредленных транзакции использован паттерн "сага хореографией". Для обмена сообщениями в рамках этого паттерна выбран брокер сообщений RabbitMQ.
2. Всего есть 4 сервиса(заказ,оплата,склад, доставка), у каждого сервиса своя БД. Общая логика работы: 
   а.запрос от клиента на создание заказа поступает в сервис заказа,
   б.сервис заказа регистрирует заказ в своей БД и передает его в сервис оплаты
   в.сервис оплаты получает заказ, создает операцию оплаты в своей БД, обновляет статус заказа на "Оплачен" и передает заказ в сервис склада
   г.сервис склада получает заказ, создает операцию резерва товаров из заказа на складе, обновляет статус заказа на "Зарезервирован" и передает заказ в сервис доставки
   д.сервис доставки получает заказ, создает операцию доставки товара по адресу из заказа и обновляет статус заказа на "Доставляется".
   е.при отмене заказа или при неудачной попытке выполнить одну их операций(оплаты/резерва/доставки) обновляется статус заказа на "Отменен" и каждым сервисом выполняются компенсирующие операции отката(оплаты/резерва/доставки)
3. Для обмена заказами между сервисами в RabbitMQ сервисом заказов создаётся один общий exchange и каждый сервис при старте подключает свою очередь к этому exchange с соответствующим ключом маршрутизации. При отправке заказа в exchange сервис указывает ключ маршрутизации, тем самым гарантируется доставка заказа из общего exchange в очередь конкретного сервиса.
4. Для обмена статусами заказов использована аналогичная схема за тем лишь исключением что ключи маршрутизации не используются, а сообщения из exchange доставляются всем сервисам. При этом обновление статуса заказа в БД выполняет только сервис заказов, так как только он ответственен за заказы и только у него есть доступ к БД с заказами.
5. Для удобства тестирования реализована отдельная таблица с историей изменения статуса заказа и отдельный GET метод /api/v1/orders/status_change_list?id=<id>. В тестах Id заказа генерируется каждый раз случайным образом, что позволяет запускать прогон коллекции тестов множество раз без появления ошибок о дублировании заказов.  

