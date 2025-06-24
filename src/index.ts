export { MyContainer } from './container'
import { getRandom } from '@cloudflare/containers'
import { Hono } from 'hono'

const app = new Hono<{ Bindings: CloudflareBindings }>()

app.get('/', (c) => c.text('Hi'))

app.post('/resize', async (c) => {
  const container = await getRandom(c.env.IMAGE_RESIZE_CONTAINER)
  return container.fetch(c.req.raw)
})

export default app
