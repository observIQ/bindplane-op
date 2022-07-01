export function classes(classes: (string | undefined)[]): string {
  return classes.filter(c => c != null).join(' ');
}
