-- Registro do aceite dos Termos de Uso e da Política de Privacidade: quem aceitou,
-- qual versão do texto e quando. Uma linha por usuário e versão; uma versão nova
-- dos textos exige novo aceite. Apagar a conta apaga o registro.
-- Contas anteriores à 000035 não têm linha e aceitam no próximo acesso.
CREATE TABLE legal_acceptances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    terms_version TEXT NOT NULL CHECK (terms_version ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'),
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, terms_version)
);
