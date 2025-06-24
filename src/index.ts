import { getRandom } from '@cloudflare/containers'
import { Hono } from 'hono'
import { cache } from 'hono/cache'
import { getExtension } from 'hono/utils/mime'
import { sha256 } from 'hono/utils/crypto'
export { MyContainer } from './container'

const app = new Hono<{ Bindings: CloudflareBindings }>()

app.get('/', (c) => {
  return c.text(
    `Image Resize API

Endpoints:
- PUT /upload - Upload an image
- GET /:key - Get/resize an image

Usage Examples:
1. Upload image: curl -X PUT -F "image=@photo.jpg" /upload
2. Get original: curl /image-key
3. Resize: curl /image-key?width=300&height=200`
  )
})

app.put('/upload', async (c) => {
  // Get the request
  const data = await c.req.parseBody<{ image: File }>()
  const body = data.image
  const type = data.image.type
  const extension = getExtension(type) ?? 'png'
  const key = (await sha256(await body.text())) + '.' + extension
  // Put the image to R2
  await c.env.BUCKET.put(key, body, { httpMetadata: { contentType: type } })
  return c.text(key)
})

app.get(
  '*',
  cache({
    cacheName: 'image-resize-container',
    cacheControl: 'max-age=3600'
  })
)

app.get('/:key', async (c) => {
  const key = c.req.param('key')
  // Get the image from R2
  const object = await c.env.BUCKET.get(key)
  if (!object) return c.notFound()
  const { width, height } = c.req.query()
  if (!width && !height) {
    return c.body(await object.arrayBuffer(), 200, {
      'Content-Type': object.httpMetadata?.contentType ?? ''
    })
  }
  // Create a request
  const data = await object.blob()
  const formData = new FormData()
  formData.append('image', data)
  const searchParams = new URLSearchParams(c.req.query())
  // Fetch a resized image from the container
  const container = await getRandom(c.env.IMAGE_RESIZE_CONTAINER)
  return container.fetch('http://localhost/resize?' + searchParams.toString(), {
    method: 'POST',
    body: formData
  })
})

export default app
