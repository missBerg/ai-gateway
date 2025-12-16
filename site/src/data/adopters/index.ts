export type Adopter = {
  name: string;
  logoUrl: string;
  url?: string;
  description?: string;
};

// Dynamically import all adopter JSON files (except template and README)
// Using require.context for Docusaurus/Webpack compatibility
const adopterFiles = (require as any).context(
  './',
  false,
  /^\.\/(?!_template|adopters)[^\/]+\.json$/
);

// Load all adopters from individual JSON files
export const adopters: Adopter[] = adopterFiles.keys().map((fileName: string) => {
  return adopterFiles(fileName) as Adopter;
});

// Sort adopters alphabetically by name
export const sortedAdopters: Adopter[] = [...adopters].sort((a, b) =>
  a.name.localeCompare(b.name)
);
