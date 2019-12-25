VERSION=0.1.6

docker build -t graphql-gateway .

docker tag graphql-gateway nandiheath/graphql-gateway:$VERSION
docker push nandiheath/graphql-gateway:$VERSION