import { Resvg } from '@resvg/resvg-js'
import { readFileSync, writeFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const svgPath = resolve(__dirname, '..', 'public', 'og.svg')
const pngPath = resolve(__dirname, '..', 'public', 'og.png')

const svg = readFileSync(svgPath, 'utf-8')

const resvg = new Resvg(svg, {
  fitTo: {
    mode: 'width',
    value: 1200,
  },
  font: {
    loadSystemFonts: true,
  },
})

const pngData = resvg.render()
const pngBuffer = pngData.asPng()

writeFileSync(pngPath, pngBuffer)
console.log(`✅ Generated ${pngPath} (${(pngBuffer.length / 1024).toFixed(1)} KB)`)
