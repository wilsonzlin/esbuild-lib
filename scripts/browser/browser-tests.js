const puppeteer = require('puppeteer')
const http = require('http')
const path = require('path')
const url = require('url')
const fs = require('fs')

const js = fs.readFileSync(path.join(__dirname, '..', '..', 'npm', 'esbuild-wasm', 'lib', 'browser.js'))
const esm = fs.readFileSync(path.join(__dirname, '..', '..', 'npm', 'esbuild-wasm', 'esm', 'browser.js'))
const wasm = fs.readFileSync(path.join(__dirname, '..', '..', 'npm', 'esbuild-wasm', 'esbuild.wasm'))

// This is converted to a string and run inside the browser
async function runAllTests({ esbuild, service }) {
  const tests = {
    async transformJS() {
      const { code } = await service.transform('1+2')
      assertStrictEqual(code, '1 + 2;\n')
    },

    async transformTS() {
      const { code } = await service.transform('1 as any + <any>2', { loader: 'ts' })
      assertStrictEqual(code, '1 + 2;\n')
    },

    async transformCSS() {
      const { code } = await service.transform('div { color: red }', { loader: 'css' })
      assertStrictEqual(code, 'div {\n  color: red;\n}\n')
    },

    async buildFib() {
      const fibonacciPlugin = {
        name: 'fib',
        setup(build) {
          build.onResolve({ filter: /^fib\((\d+)\)/ }, args => {
            return { path: args.path, namespace: 'fib' }
          })
          build.onLoad({ filter: /^fib\((\d+)\)/, namespace: 'fib' }, args => {
            let match = /^fib\((\d+)\)/.exec(args.path), n = +match[1]
            let contents = n < 2 ? `export default ${n}` : `
              import n1 from 'fib(${n - 1}) ${args.path}'
              import n2 from 'fib(${n - 2}) ${args.path}'
              export default n1 + n2`
            return { contents }
          })
        },
      }
      const result = await service.build({
        stdin: {
          contents: `
            import x from 'fib(10)'
            return x
          `,
        },
        format: 'cjs',
        bundle: true,
        plugins: [fibonacciPlugin],
      })
      assertStrictEqual(result.outputFiles.length, 1)
      assertStrictEqual(result.outputFiles[0].path, '<stdout>')
      const code = result.outputFiles[0].text
      const fib10 = new Function(code)()
      assertStrictEqual(fib10, 55)
    },

    async serve() {
      expectThrownError(service.serve, 'The "serve" API only works in node')
    },

    async esbuildBuild() {
      expectThrownError(esbuild.build, 'The "build" API only works in node')
    },

    async esbuildTransform() {
      expectThrownError(esbuild.transform, 'The "transform" API only works in node')
    },

    async esbuildBuildSync() {
      expectThrownError(esbuild.buildSync, 'The "buildSync" API only works in node')
    },

    async esbuildTransformSync() {
      expectThrownError(esbuild.transformSync, 'The "transformSync" API only works in node')
    },
  }

  function expectThrownError(fn, err) {
    try {
      fn()
      throw new Error('Expected an error to be thrown')
    } catch (e) {
      assertStrictEqual(e.message, err)
    }
  }

  function assertStrictEqual(a, b) {
    if (a !== b) {
      throw new Error(`Assertion failed:
  Expected: ${JSON.stringify(a)}
  Observed: ${JSON.stringify(b)}`);
    }
  }

  async function runTest(test) {
    try {
      await tests[test]()
    } catch (e) {
      testFail(`[${test}] ` + (e && e.message || e))
    }
  }

  const promises = []
  for (const test in tests) {
    promises.push(runTest(test))
  }
  await Promise.all(promises)
}

let pages = {
  iife: `
    <script src="/lib/esbuild.js"></script>
    <script>
      testStart = function() {
        esbuild.startService({ wasmURL: '/esbuild.wasm' }).then(service => {
          return (${runAllTests})({ esbuild, service })
        }).then(() => {
          testDone()
        }).catch(e => {
          testFail('' + (e && e.stack || e))
          testDone()
        })
      }
    </script>
  `,
  esm: `
    <script type="module">
      import * as esbuild from '/esm/esbuild.js'
      window.testStart = function() {
        esbuild.startService({ wasmURL: '/esbuild.wasm' }).then(service => {
          return (${runAllTests})({ esbuild, service })
        }).then(() => {
          testDone()
        }).catch(e => {
          testFail('' + (e && e.stack || e))
          testDone()
        })
      }
    </script>
  `,
}

const server = http.createServer((req, res) => {
  if (req.method === 'GET' && req.url) {
    if (req.url === '/lib/esbuild.js') {
      res.writeHead(200, { 'Content-Type': 'application/javascript' })
      res.end(js)
      return
    }

    if (req.url === '/esm/esbuild.js') {
      res.writeHead(200, { 'Content-Type': 'application/javascript' })
      res.end(esm)
      return
    }

    if (req.url === '/esbuild.wasm') {
      res.writeHead(200, { 'Content-Type': 'application/wasm' })
      res.end(wasm)
      return
    }

    if (req.url.startsWith('/page/')) {
      let key = req.url.slice('/page/'.length)
      if (Object.prototype.hasOwnProperty.call(pages, key)) {
        res.writeHead(200, { 'Content-Type': 'text/html' })
        res.end(`
          <!doctype html>
          <meta charset="utf8">
          ${pages[key]}
        `)
        return
      }
    }
  }

  console.log(`[http] ${req.method} ${req.url}`)
  res.writeHead(404)
  res.end()
})

server.listen()
const { address, port } = server.address()
const serverURL = url.format({ protocol: 'http', hostname: address, port })
console.log(`[http] listening on ${serverURL}`)

async function main() {
  const browser = await puppeteer.launch()
  const promises = []
  let allTestsPassed = true

  async function runPage(key) {
    try {
      const page = await browser.newPage()
      page.on('console', obj => console.log(`[console.${obj.type()}] ${obj.text()}`))
      page.exposeFunction('testFail', error => {
        console.log(`❌ ${error}`)
        allTestsPassed = false
      })
      let testDone = new Promise(resolve => {
        page.exposeFunction('testDone', resolve)
      })
      await page.goto(`${serverURL}/page/${key}`, { waitUntil: 'domcontentloaded' })
      await page.evaluate('testStart()')
      await testDone
      await page.close()
    } catch (e) {
      allTestsPassed = false
      console.log(`❌ ${key}: ${e && e.message || e}`)
    }
  }

  for (let key in pages) {
    promises.push(runPage(key))
  }

  await Promise.all(promises)
  await browser.close()
  server.close()

  if (!allTestsPassed) {
    console.error(`❌ browser test failed`)
    process.exit(1)
  } else {
    console.log(`✅ browser test passed`)
  }
}

main().catch(error => setTimeout(() => { throw error }))
