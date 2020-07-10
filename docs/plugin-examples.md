# Example plugins

## Svelte plugin

```js
let svelte = require('svelte/compiler')

// import value from './example.svelte'
let sveltePlugin = plugin => {
  plugin.setName('svelte');
  plugin.addLoader({ filter: /\.svelte$/ }, args => {
    let source = fs.readFileSync(args.path, 'utf8')
    return { contents: svelte.compile(source).js.code }
  })
}
```

## PnP Yarn plugin

```js
let { PosixFS, ZipOpenFS } = require(`@yarnpkg/fslib`)
let libzip = require(`@yarnpkg/libzip`).getLibzipSync()
let zipOpenFs = new ZipOpenFS({ libzip })
let crossFs = new PosixFS(zipOpenFs)
let pnpapi = require('pnpapi')

// More info: https://classic.yarnpkg.com/en/docs/pnp/
let pnpYarnPlugin = plugin => {
  plugin.setName('pnp')
  plugin.addResolver({ filter: /.*/ }, args => {
    return { path: pnpapi.resolveRequest(args.path, args.importDir + '/') }
  })
  plugin.addLoader({ filter: /.*/ }, args => {
    return { contents: crossFs.readFileSync(args.path), loader: 'default' }
  })
}
```

### YAML plugin

```js
let YAML = require('js-yaml')
let util = require('util')
let fs = require('fs')

// import value from './example.yaml'
let yamlLoader = plugin => {
  plugin.setName('yaml-loader')
  plugin.addLoader({ filter: /\.ya?ml$/ }, async (args) => {
    let source = await util.promisify(fs.readFile)(args.path, 'utf8')
    try {
      let contents = JSON.stringify(YAML.safeLoad(source), null, 2)
      return { contents, loader: 'json' }
    } catch (e) {
      return {
        errors: [{
          text: (e && e.reason) || (e && e.message) || e,
          location: e.mark && {
            line: e.mark.line,
            column: e.mark.column,
            lineText: source.split(/\r\n|\r|\n/g)[e.mark.line],
          },
        }],
      }
    }
  })
}
```

### SVG optimizer plugin

```js
let SVGO = require('svgo')
let util = require('util')
let fs = require('fs')

// import value from './example.svg'
let svgOptimizer = plugin => {
  plugin.setName('svg-optimizer')
  plugin.addLoader({ filter: /\.svg$/ }, async (args) => {
    let source = await util.promisify(fs.readFile)(args.path, 'utf8')
    try {
      let { data: contents } = await new SVGO().optimize(source)
      return { contents, loader: 'text' }
    } catch (e) {
      let match = /^(.*)\nLine: (\d+)\nColumn: (\d+)\nChar: (.*)$/.exec(e + '')
      return {
        errors: [{
          text: match ? match[1] : (e && e.message) || e,
          location: match && {
            line: +match[2],
            column: +match[3] - 1,
            length: match[4].length,
            lineText: source.split(/\r\n|\r|\n/g)[+match[2]],
          },
        }],
      }
    }
  })
}
```

### CoffeeScript plugin

```js
// import value from './example.coffee'
let CoffeeScript = require('coffeescript')
let util = require('util')
let fs = require('fs')

let coffeeLoader = plugin => {
  plugin.setName('coffee-loader')
  plugin.addLoader({ filter: /\.coffee$/ }, async (args) => {
    let source = await util.promisify(fs.readFile)(args.path, 'utf8')
    try {
      return { contents: CoffeeScript.compile(source) }
    } catch (e) {
      return {
        errors: [{
          text: (e && e.message) || e,
          location: e.location && {
            line: e.location.first_line,
            column: e.location.first_column,
            length: e.location.last_column - e.location.first_column + 1,
            lineText: source.split(/\r\n|\r|\n/g)[e.location.first_line],
          },
        }],
      }
    }
  })
}
```

## HTTP plugin

```js
// import value from 'https://www.google.com'
let https = require('https')
let http = require('http')

let httpLoader = plugin => {
  plugin.setName('http')
  plugin.addResolver({ filter: /^https?:\/\// }, args => {
    return { path: args.path, namespace: 'http' }
  })
  plugin.addLoader({ filter: /^https?:\/\//, namespace: 'http' }, async (args) => {
    let contents = await new Promise((resolve, reject) => {
      let lib = args.path.startsWith('https') ? https : http
      lib.get(args.path, res => {
        let chunks = []
        res.on('data', chunk => chunks.push(chunk))
        res.on('end', () => resolve(Buffer.concat(chunks)))
      }).on('error', reject)
    })
    return { contents, loader: 'text' }
  })
}
```

## Fibonacci plugin

```js
// import value from 'fib(10)'
let fibonacciLoader = plugin => {
  plugin.setName('fibonacci')
  plugin.addResolver({ filter: /^fib\((\d+)\)/ }, args => {
    return { path: args.path, namespace: 'fibonacci' }
  })
  plugin.addLoader({ filter: /^fib\((\d+)\)/, namespace: 'fibonacci' }, args => {
    let match = /^fib\((\d+)\)/.exec(args.path), n = +match[1]
    let contents = n < 2 ? `export default ${n}` : `
      import n1 from 'fib(${n - 1}) ${args.path}'
      import n2 from 'fib(${n - 2}) ${args.path}'
      export default n1 + n2`
    return { contents }
  })
}
```
