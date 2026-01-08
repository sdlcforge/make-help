#!/usr/bin/env node

/**
 * Badge generator using badge-maker.
 *
 * Usage:
 *   node badger.mjs coverage <percent> <output-file>
 *   node badger.mjs license <output-file>
 *
 * Examples:
 *   node badger.mjs coverage 85.5 ./docs/assets/coverage-badge.svg
 *   node badger.mjs license ./docs/assets/license-badge.svg
 */

import { makeBadge } from 'badge-maker'
import { writeFileSync, mkdirSync } from 'fs'
import { dirname } from 'path'

const args = process.argv.slice(2)
const command = args[0]

function showUsage() {
  console.error('Usage:')
  console.error('  node badger.mjs coverage <percent> <output-file>')
  console.error('  node badger.mjs license <output-file>')
  console.error('')
  console.error('Examples:')
  console.error('  node badger.mjs coverage 85.5 ./docs/assets/coverage-badge.svg')
  console.error('  node badger.mjs license ./docs/assets/license-badge.svg')
  process.exit(1)
}

/**
 * Write badge SVG to file, creating directories as needed.
 */
function writeBadge(badge, outputFile) {
  mkdirSync(dirname(outputFile), { recursive: true })
  writeFileSync(outputFile, badge)
}

/**
 * Calculate badge color based on coverage percentage.
 * Uses HSL color space for smooth red-to-green gradient.
 *
 * @param {number} coverage - Coverage percentage (0-100)
 * @returns {{ color: string, hue: number }} HSL color string and hue value
 */
function getCoverageColor(coverage) {
  // Normalize to 0-1 range
  const normalized = coverage / 100

  // Coverage <= 50% stays red
  // Coverage > 50% transitions toward green
  let colorshift
  if (normalized <= 0.5) {
    colorshift = 1 // Full red
  } else {
    // Linear transition from red (at 50%) to green (at 100%)
    colorshift = 1 - (normalized - 0.5) * 2
  }

  // HSL hue: 0 = red, 120 = green
  const hue = Math.round((1 - colorshift) * 120)

  // Use 100% saturation and 35% lightness for readability
  return {
    color: `hsl(${hue},100%,35%)`,
    hue
  }
}

/**
 * Generate coverage badge.
 */
function generateCoverageBadge(percent, outputFile) {
  if (isNaN(percent) || percent < 0 || percent > 100) {
    console.error(`Invalid coverage percentage: ${percent}`)
    console.error('Must be a number between 0 and 100')
    process.exit(1)
  }

  const coverageRounded = Math.round(percent * 10) / 10
  const { color, hue } = getCoverageColor(percent)

  const badge = makeBadge({
    label: 'coverage',
    message: `${coverageRounded}%`,
    color,
    style: 'flat'
  })

  writeBadge(badge, outputFile)
  console.log(`Coverage badge generated: ${coverageRounded}% (hue: ${hue}) -> ${outputFile}`)
}

/**
 * Generate license badge.
 * Uses Apache Software Foundation brand color (#7C297D) from their style guide.
 */
function generateLicenseBadge(outputFile) {
  // ASF main logo color from their style guide
  const ASF_COLOR = '#7C297D'

  const badge = makeBadge({
    label: 'license',
    message: 'Apache 2.0',
    color: ASF_COLOR,
    style: 'flat'
  })

  writeBadge(badge, outputFile)
  console.log(`License badge generated: Apache 2.0 (${ASF_COLOR}) -> ${outputFile}`)
}

// Main command dispatch
if (!command) {
  showUsage()
}

switch (command) {
  case 'coverage': {
    if (args.length < 3) {
      console.error('Error: coverage command requires <percent> and <output-file>')
      showUsage()
    }
    const percent = parseFloat(args[1])
    const outputFile = args[2]
    generateCoverageBadge(percent, outputFile)
    break
  }

  case 'license': {
    if (args.length < 2) {
      console.error('Error: license command requires <output-file>')
      showUsage()
    }
    const outputFile = args[1]
    generateLicenseBadge(outputFile)
    break
  }

  default:
    console.error(`Unknown command: ${command}`)
    showUsage()
}
