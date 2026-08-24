# Kashi Project Guidelines

## MDI Icons: Always Use Treeshaken Imports

When using Material Design Icons (MDI) in the UI, always use treeshaken imports from `@mdi/js` to minimize bundle size.

**Pattern:**
1. Import only the specific icons you need from `@mdi/js` at the top of `ui/src/plugins/icons.js`
2. Add the icon to the `iconPaths` object, mapping kebab-case names to the imported constant
3. Use the icon in templates with the format `mdi-icon-name`

**Example:**
```javascript
// In ui/src/plugins/icons.js
import { mdiAccount, mdiPlus } from '@mdi/js'

const iconPaths = {
  'mdi-account': mdiAccount,
  'mdi-plus': mdiPlus,
}
```

Then use in templates:
```vue
<v-icon icon="mdi-account" />
```

Never use the full MDI icon set or `mdi-js` default export — always treeshake specific icons.
