import type { Configuration } from 'webpack'
import path from 'path'

import grafanaConfig, { type Env } from './.config/webpack/webpack.config'

const config = async (env: Env): Promise<Configuration> => {
  const baseConfig = await grafanaConfig(env)

  return {
    ...baseConfig,
    // Limit entries to root plugin only (avoid nested scaffolds in repo)
    entry: {
      module: path.resolve(process.cwd(), 'src', 'module.ts'),
    },
  }
}

export default config
