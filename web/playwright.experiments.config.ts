import {defineConfig} from '@playwright/test'
import browserConfig from './playwright.config'

// Rejected/unadopted prototypes are executable diagnostics, not default product
// paths. Keep their strict comparisons intact and their failures reviewable.
export default defineConfig({...browserConfig,testDir:'./tests/experiments',outputDir:'experiment-test-results'})
