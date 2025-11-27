import path from 'node:path'

const addHelp = async ({ makefilePath = path.join(process.cwd(), 'Makefile') }) => {
    process.stdout.write(`Adding help to ${makefilePath}\n`) // remove me
}

export default addHelp