import { describe, expect, it } from 'vitest';
import { APP_VERSION, UPDATE_NOTES, updateNotes } from './release';
import { UPDATE_NOTES_EN } from './release.en';

describe('notas de versão', () => {
  it('o inglês tem uma nota para cada nota em português, na mesma ordem', () => {
    expect(UPDATE_NOTES_EN.map((note) => note.version)).toEqual(UPDATE_NOTES.map((note) => note.version));
  });

  it('a versão atual tem nota nos dois idiomas', () => {
    expect(updateNotes('pt').some((note) => note.version === APP_VERSION)).toBe(true);
    expect(updateNotes('en').some((note) => note.version === APP_VERSION)).toBe(true);
  });
});
