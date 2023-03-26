all:

CMD=main
NAME=exe

build: css
	CGO_ENABLED=0 go build -o $(NAME) ./$(CMD).go

serve:
	$(NAME) --local

space-build-linux-amd64: css
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o micro/$(NAME) ./$(CMD).go
	ls -al micro/$(NAME)
	xz -9 -f micro/$(NAME)
	ls -al micro/$(NAME)*

deploy: space-build-linux-amd64 push

push:
	(cd micro && spacex push --verbose)

css:
	tailwindcss -m -i ./tailwind.css -o ./site/style.css

watch:
	npx watch "make css" ./site

ui-build:
	npm run build

ui-serve:
	npm run dev

ui-preview:
	npm run preview
