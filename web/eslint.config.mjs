import { defineConfig } from "eslint/config";
import { FlatCompat } from "@eslint/eslintrc";
import nextCoreWebVitals from "eslint-config-next/core-web-vitals.js";
import nextTypescript from "eslint-config-next/typescript.js";
import i18Next from "eslint-plugin-i18next";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

export default defineConfig([
  {
    ignores: ["**/.next/**", "**/node_modules/**", "**/next-env.d.ts"],
  },
  {
    linterOptions: {
      noInlineConfig: true,
    },
  },
  ...compat.extends(...nextCoreWebVitals.extends, ...nextTypescript.extends),
  {
    files: ["**/*.{ts,tsx}"],
    plugins: {
      i18next: i18Next,
    },
    rules: {
      "i18next/no-literal-string": [
        "error",
        {
          markupOnly: false,
          ignoreAttribute: [
            "data",
            "aria-label",
            "dir",
            "lang",
            "role",
            "tabIndex",
            "style",
            "className",
            "onChange",
            "onClick",
            "onSubmit",
            "onBlur",
            "onFocus",
            "onKeyDown",
            "onKeyUp",
            "onMouseEnter",
            "onMouseLeave",
            "onScroll",
            "type",
            "name",
            "value",
            "defaultValue",
            "placeholder",
            "autoComplete",
            "inputMode",
            "accept",
            "method",
            "action",
            "href",
            "target",
            "rel",
            "src",
            "alt",
            "width",
            "height",
            "viewBox",
            "fill",
            "stroke",
            "d",
            "rx",
            "ry",
            "xmlns",
            "key",
            "index",
          ],
          ignoreComponent: [
            "AvatarFallback",
            "CardTitle",
            "CardDescription",
            "CardHeader",
            "CardContent",
            "CardFooter",
            "SelectValue",
            "Badge",
            "TableHead",
            "TableCell",
            "TableRow",
            "TableHeader",
            "TableBody",
            "TabsList",
            "TabsTrigger",
            "TabsContent",
            "SelectItem",
            "SelectGroup",
            "Option",
            "Separator",
            "Slider",
            "Switch",
            "Progress",
            "TooltipContent",
            "DialogTitle",
            "DialogDescription",
            "DialogContent",
            "SheetTitle",
            "SheetDescription",
            "SheetContent",
            "PopoverContent",
            "CommandInput",
            "CommandEmpty",
            "CommandGroup",
            "CommandItem",
            "CommandList",
            "CommandDialog",
            "CommandSeparator",
          ],
        },
      ],
    },
  },
]);
