// Type declaration for webpack's require.context
declare const require: {
  context: (
    directory: string,
    useSubdirectories: boolean,
    regExp: RegExp
  ) => {
    keys(): string[];
    (id: string): any;
  };
};

// Dynamically import all adopter JSON files
function importAll(r: ReturnType<typeof require.context>) {
  return r.keys().map(r);
}

// Create a context that matches all JSON files except this index file
const adopterContext = require.context('./', false, /^\.\/(?!index).*\.json$/);

// Import all adopter JSON files dynamically
export const adopters = importAll(adopterContext);

export type Adopter = {
  name: string;
  logoUrl: string;
  url?: string;
  description?: string;
};

// Sort adopters alphabetically by name
export const sortedAdopters = adopters.sort((a, b) => a.name.localeCompare(b.name));
