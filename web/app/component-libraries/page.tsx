import type { Metadata } from "next";

import ShowcaseAvatar from "./_containers/ShowcaseAvatar";
import ShowcaseBadge from "./_containers/ShowcaseBadge";
import ShowcaseButton from "./_containers/ShowcaseButton";
import ShowcaseMenu from "./_containers/ShowcaseMenu";

export const metadata: Metadata = {
  title: "Component Libraries",
  description: "Every variant of the Wongnok base components, with usage.",
};

const SECTIONS = [
  { href: "#button", label: "Button" },
  { href: "#badge", label: "Badge" },
  { href: "#avatar", label: "Avatar" },
  { href: "#menu", label: "Dropdown menu" },
];

const ComponentLibraries = () => {
  return (
    <div className="flex flex-1 flex-col bg-background">
      <main className="mx-auto flex w-full max-w-5xl flex-col gap-16 px-6 py-16">
        <header className="flex flex-col gap-4">
          <h1 className="wongnok-text-h1 text-foreground">
            Component Libraries
          </h1>
          <p className="wongnok-text-body max-w-xl text-muted-foreground">
            Every variant of the Wongnok base components, with the JSX that
            produces it.
          </p>
          <nav className="flex flex-wrap gap-2">
            {SECTIONS.map((section) => (
              <a
                key={section.href}
                href={section.href}
                className="wongnok-text-label rounded-full border border-border bg-card px-3 py-1 text-muted-foreground transition-colors hover:border-primary hover:text-primary"
              >
                {section.label}
              </a>
            ))}
          </nav>
        </header>

        <ShowcaseButton />
        <ShowcaseBadge />
        <ShowcaseAvatar />
        <ShowcaseMenu />
      </main>
    </div>
  );
};

export default ComponentLibraries;
