#!/usr/bin/env node

/**
 * Generates a coverage badge SVG using badge-maker.
 *
 * Usage: node make-coverage-badge.js <coverage-percent> <output-file>
 * Example: node make-coverage-badge.js 85.5 ./docs/assets/coverage-badge.svg
 *
 * Color scheme:
 * - 50% or less: red
 * - 50% to 100%: gradient from red to green (using HSL color space)
 */

import { makeBadge } from 'badge-maker'
import { writeFileSync, mkdirSync } from 'fs'
import { dirname } from 'path'

const args = process.argv.slice(2)

if (args.length < 2) {
  console.error('Usage: node make-coverage-badge.js <coverage-percent> <output-file>')
  console.error('Example: node make-coverage-badge.js 85.5 ./docs/assets/coverage-badge.svg')
  process.exit(1)
}

const coveragePercent = parseFloat(args[0])
const outputFile = args[1]

if (isNaN(coveragePercent) || coveragePercent < 0 || coveragePercent > 100) {
  console.error(`Invalid coverage percentage: ${args[0]}`)
  console.error('Must be a number between 0 and 100')
  process.exit(1)
}

/**
 * Calculate badge color based on coverage percentage.
 * Uses HSL color space for smooth red-to-green gradient.
 *
 * @param {number} coverage - Coverage percentage (0-100)
 * @returns {string} HSL color string
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
  return `hsl(${hue},100%,35%)`
}

// Round coverage to one decimal place for display
const coverageRounded = Math.round(coveragePercent * 10) / 10

// Generate the badge
const badge = makeBadge({
  label: 'coverage',
  message: `${coverageRounded}%`,
  color: getCoverageColor(coveragePercent),
  style: 'flat'
})

// Ensure output directory exists
mkdirSync(dirname(outputFile), { recursive: true })

// Write the badge
writeFileSync(outputFile, badge)

// Calculate hue for display (same logic as getCoverageColor)
const normalized = coveragePercent / 100
const colorshift = normalized <= 0.5 ? 1 : 1 - (normalized - 0.5) * 2
const hue = Math.round((1 - colorshift) * 120)

console.log(`Coverage badge generated: ${coverageRounded}% (hue: ${hue}) -> ${outputFile}`)
