const { installForTests } = require('./esbuild')
const rimraf = require('rimraf')
const assert = require('assert')
const path = require('path')
const util = require('util')
const fs = require('fs')

const readFileAsync = util.promisify(fs.readFile)
const writeFileAsync = util.promisify(fs.writeFile)
const mkdirAsync = util.promisify(fs.mkdir)

const repoDir = path.dirname(__dirname)
const rootTestDir = path.join(repoDir, 'scripts', '.plugin-tests')

let pluginTests = {
  async basicLoader({ esbuild, testDir }) {
    const input = path.join(testDir, 'in.js')
    const custom = path.join(testDir, 'example.custom')
    const output = path.join(testDir, 'out.js')
    await writeFileAsync(input, `
      import x from './example.custom'
      export default x
    `)
    await writeFileAsync(custom, ``)
    await esbuild.build({
      entryPoints: [input], bundle: true, outfile: output, format: 'cjs', plugins: [
        plugin => {
          plugin.setName('name')
          plugin.addLoader({ filter: /\.custom$/ }, args => {
            assert.strictEqual(args.path, custom)
            return { contents: 'this is custom', loader: 'text' }
          })
        },
      ],
    })
    const result = require(output)
    assert.strictEqual(result.default, 'this is custom')
  },

  async basicResolver({ esbuild, testDir }) {
    const input = path.join(testDir, 'in.js')
    const custom = path.join(testDir, 'example.txt')
    const output = path.join(testDir, 'out.js')
    await writeFileAsync(input, `
      import x from 'test'
      export default x
    `)
    await writeFileAsync(custom, `example text`)
    await esbuild.build({
      entryPoints: [input], bundle: true, outfile: output, format: 'cjs', plugins: [
        plugin => {
          plugin.setName('name')
          plugin.addResolver({ filter: /^test$/ }, args => {
            assert.strictEqual(args.path, 'test')
            return { path: custom }
          })
        },
      ],
    })
    const result = require(output)
    assert.strictEqual(result.default, 'example text')
  },

  async fibonacciResolverMemoized({ esbuild, testDir }) {
    const input = path.join(testDir, 'in.js')
    const output = path.join(testDir, 'out.js')
    await writeFileAsync(input, `
      import x from 'fib(10)'
      export default x
    `)
    await esbuild.build({
      entryPoints: [input], bundle: true, outfile: output, format: 'cjs', plugins: [
        plugin => {
          plugin.setName('name')
          plugin.addResolver({ filter: /^fib\((\d+)\)$/ }, args => {
            return { path: args.path, namespace: 'fib' }
          })
          plugin.addLoader({ filter: /^fib\((\d+)\)$/, namespace: 'fib' }, args => {
            let match = /^fib\((\d+)\)$/.exec(args.path), n = +match[1]
            let contents = n < 2 ? `export default ${n}` : `
              import n1 from 'fib(${n - 1})'
              import n2 from 'fib(${n - 2})'
              export default n1 + n2`
            return { contents }
          })
        },
      ],
    })
    const result = require(output)
    assert.strictEqual(result.default, 55)
  },

  async fibonacciResolverNotMemoized({ esbuild, testDir }) {
    const input = path.join(testDir, 'in.js')
    const output = path.join(testDir, 'out.js')
    await writeFileAsync(input, `
      import x from 'fib(10)'
      export default x
    `)
    await esbuild.build({
      entryPoints: [input], bundle: true, outfile: output, format: 'cjs', plugins: [
        plugin => {
          plugin.setName('name')
          plugin.addResolver({ filter: /^fib\((\d+)\)/ }, args => {
            return { path: args.path, namespace: 'fib' }
          })
          plugin.addLoader({ filter: /^fib\((\d+)\)/, namespace: 'fib' }, args => {
            let match = /^fib\((\d+)\)/.exec(args.path), n = +match[1]
            let contents = n < 2 ? `export default ${n}` : `
              import n1 from 'fib(${n - 1}) ${args.path}'
              import n2 from 'fib(${n - 2}) ${args.path}'
              export default n1 + n2`
            return { contents }
          })
        },
      ],
    })
    const result = require(output)
    assert.strictEqual(result.default, 55)
  },

  async resolversCalledInSequence({ esbuild, testDir }) {
    const input = path.join(testDir, 'in.js')
    const nested = path.join(testDir, 'nested.js')
    const output = path.join(testDir, 'out.js')
    await writeFileAsync(input, `
      import x from 'test'
      export default x
    `)
    await writeFileAsync(nested, `
      export default 123
    `)
    let trace = []
    await esbuild.build({
      entryPoints: [input], bundle: true, outfile: output, format: 'cjs', plugins: [
        plugin => {
          plugin.setName('plugin1')
          plugin.addResolver({ filter: /^.*$/ }, () => trace.push('called first'))
        },
        plugin => {
          plugin.setName('plugin2')
          plugin.addResolver({ filter: /^ignore me$/ }, () => trace.push('not called'))
        },
        plugin => {
          plugin.setName('plugin3')
          plugin.addResolver({ filter: /^.*$/ }, () => {
            trace.push('called second')
            return { path: nested }
          })
        },
        plugin => {
          plugin.setName('plugin4')
          plugin.addResolver({ filter: /^.*$/ }, () => trace.push('not called'))
        },
      ],
    })
    const result = require(output)
    assert.strictEqual(result.default, 123)
    assert.deepStrictEqual(trace, [
      'called first',
      'called second',
    ])
  },

  async loadersCalledInSequence({ esbuild, testDir }) {
    const input = path.join(testDir, 'in.js')
    const nested = path.join(testDir, 'nested.js')
    const output = path.join(testDir, 'out.js')
    await writeFileAsync(input, `
      import x from './nested.js'
      export default x
    `)
    await writeFileAsync(nested, `
      export default 123
    `)
    let trace = []
    await esbuild.build({
      entryPoints: [input], bundle: true, outfile: output, format: 'cjs', plugins: [
        plugin => {
          plugin.setName('plugin1')
          plugin.addLoader({ filter: /^.*$/ }, () => trace.push('called first'))
        },
        plugin => {
          plugin.setName('plugin2')
          plugin.addLoader({ filter: /^.*$/, namespace: 'ignore-me' }, () => trace.push('not called'))
        },
        plugin => {
          plugin.setName('plugin3')
          plugin.addLoader({ filter: /^.*$/, namespace: 'file' }, () => {
            trace.push('called second')
            return { contents: 'export default "abc"' }
          })
        },
        plugin => {
          plugin.setName('plugin4')
          plugin.addLoader({ filter: /^.*$/, namespace: 'file' }, () => trace.push('not called'))
        },
      ],
    })
    const result = require(output)
    assert.strictEqual(result.default, 'abc')
    assert.deepStrictEqual(trace, [
      'called first',
      'called second',
    ])
  },
}

async function main() {
  // Start the esbuild service
  const esbuild = installForTests(rootTestDir)
  const service = await esbuild.startService()

  // Run all tests concurrently
  const runTest = async ([name, fn]) => {
    let testDir = path.join(rootTestDir, name)
    try {
      await mkdirAsync(testDir)
      await fn({ esbuild, service, testDir })
      rimraf.sync(testDir, { disableGlob: true })
      return true
    } catch (e) {
      console.error(`❌ ${name}: ${e && e.message || e}`)
      return false
    }
  }
  const tests = Object.entries(pluginTests)
  const allTestsPassed = (await Promise.all(tests.map(runTest))).every(success => success)

  // Clean up test output
  service.stop()

  if (!allTestsPassed) {
    console.error(`❌ plugin tests failed`)
    process.exit(1)
  } else {
    console.log(`✅ plugin tests passed`)
    rimraf.sync(rootTestDir, { disableGlob: true })
  }
}

main().catch(e => setTimeout(() => { throw e }))
