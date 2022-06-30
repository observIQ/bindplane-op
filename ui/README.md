# BindPlane UI

This directory contains the UI portion for BindPlane which is a single page react app.

## Development

To bootstrap the repo to run in development mode - run the following commands in the project root directory.

`make install-ui` - This will run an `npm install` in the ui directory.
`make prep` - This will create a file `ui/build/index.html` needed for the server to compile.

Now for a development environment you can run:

```sh
make dev
```

in the root project directory. This

1. Starts the react app with hot reloading with npm start
2. Starts the BindPlane server with `go run cmd/bindplane main.go serve`

For development navigate to `http:localhost:3000`

## GraphQL Subscriptions on Chrome

Its a known issue that the websocket connections needed for live updates for agents and configurations does not work for Chrome. It is known to work on Safari and Firefox.

## Building

To build the UI portion in the root project directory run

```sh
make ui-build
```

This will

1. Clean install node modules - `cd ui && npm ci`
2. Build the react app `npm run build`

To embed the static files in the bindplane build simply run `make build`.

Now BindPlane UI will be available at the BindPlane server url.

## Testing

To run tests on the UI from the root directory run

```
make ui-test
```

## Other Available Scripts

In the ui directory, you can run:

### `npm run eject`

This is left over from bootstrapping the project with `create-react-app`.

**Note: this is a one-way operation. Once you `eject`, you can’t go back!**

If you aren’t satisfied with the build tool and configuration choices, you can `eject` at any time. This command will remove the single build dependency from your project.

Instead, it will copy all the configuration files and the transitive dependencies (webpack, Babel, ESLint, etc) right into your project so you have full control over them. All of the commands except `eject` will still work, but they will point to the copied scripts so you can tweak them. At this point you’re on your own.

You don’t have to ever use `eject`. The curated feature set is suitable for small and middle deployments, and you shouldn’t feel obligated to use this feature. However we understand that this tool wouldn’t be useful if you couldn’t customize it when you are ready for it.
