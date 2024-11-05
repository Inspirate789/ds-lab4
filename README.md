# Лабораторная работа #4

![GitHub Classroom Workflow](../../workflows/GitHub%20Classroom%20Workflow/badge.svg?branch=master)

## Deploy to Cloud

### Формулировка

На базе [Лабораторной работы #2](https://github.com/bmstu-rsoi/lab2-template) выполнить деплой приложения в managed
кластер k8s.

### Требования

1. Скопировать исходный код из ЛР #2 в проект.
2. Развернуть руками свой Managed Kubernetes Cluster, настроить Ingress Controller (для публикации сервисов наружу можно
   использовать _только_ Ingress).
4. Собрать и опубликовать образы docker в [Docker Registry](https://hub.docker.com/).
5. Описать манифесты для деплоя в виде [helm charts](https://helm.sh/docs/topics/charts/), они должен быть универсальным
   для всех сервисов и отличаться лишь набором параметров запуска.
6. В кластере k8s можно использовать один физический instance базы, но каждый сервис должен работать только со своей
   виртуальной базой данных. Задеплоить базу в кластер можно руками, либо использовать уже готовый helm chart.
7. Код хранить на Github, для сборки использовать Github Actions.
8. Для автоматических прогонов тестов в файле [autograding.json](.github/classroom/autograding.json)
   и [classroom.yml](.github/workflows/classroom.yml) заменить `<variant>` на ваш вариант.
9. В [classroom.yml](.github/workflows/classroom.yml) дописать шаги:
    1. сборка приложения;
    2. сборка и публикация образа docker (можно использовать `docker compose build`, `docker compose push`);
    3. деплой каждого сервиса в кластер k8s.

### Пояснения

Т.к. развертывание полноценного кластера на виртуальным машинах очень сложный процесс, можно использовать Managed
Kubernetes Cluster, т.е. готовый кластер k8s, предоставляемый сторонней платформой, например:

* [Digital Ocean](https://www.digitalocean.com/products/kubernetes/)
* [Yandex Cloud](https://cloud.yandex.ru/services/managed-kubernetes)
* [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine)
* [AWS](https://aws.amazon.com/ru/eks/)

Платформ, которые предоставляют Kubernetes as a Service большое количество, вы можете сами исследовать рынок и выбрать
другого провайдера услуг. Большинство провайдеров имеют бесплатный триальный период или денежный грант.

Для создания кластера достаточно 2-3 worker ноды 2Gb, 1CPU.

### Прием задания

1. При получении задания у вас создается fork этого репозитория для вашего пользователя.
2. После того как все тесты успешно завершатся, в Github Classroom на Dashboard будет отмечено успешное выполнение
   тестов.

### Варианты заданий

Распределение вариантов заданий аналогично [ЛР #2](https://github.com/bmstu-rsoi/lab2-template).

1. [Flight Booking System](v1/README.md)
2. [Hotels Booking System](v2/README.md)
3. [Car Rental System](v3/README.md)
4. [Library System](v4/README.md)

## Car Rental System

Система предоставляет пользователю возможность забронировать автомобиль на выбранные даты.

Описание в формате [OpenAPI](%5Binst%5D%5Bv3%5D%20Car%20Rental%20System.yml).
