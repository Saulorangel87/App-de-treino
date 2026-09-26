import { useEffect, useMemo, useState } from 'react';
import { apiErrorMessage, isUnauthorized, apiRequest } from '@/lib/api';
import {
  EXCLUDED_DISCIPLINES,
  TRAINING_STATUSES,
  initialAvailability,
  initialCyclingContext,
  initialProfile,
  stepCopy,
  type Availability,
  type CyclingContext,
  type Goal,
  type Limitation,
  type Onboarding,
  type Profile,
  type ProfileForm,
  type Questionnaire,
  type User,
} from './profile-model';

export function useProfileForm() {
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<ProfileForm>(initialProfile);
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [hasLimitation, setHasLimitation] = useState(false);
  const [limitation, setLimitation] = useState({
    kind: 'pain',
    description: '',
    location: '',
    intensity: '',
    aggravating_movement: '',
    started_on: '',
    symptoms_during_after: [] as string[],
    medical_restriction: false,
    recent_surgery: false,
    exercise_prohibited: false,
    condition_affecting_exercise: false,
    professional_clearance_recommended: false,
  });
  const [primaryGoal, setPrimaryGoal] = useState({
    goal_type: 'health',
    target_date: '',
    details: '',
  });
  const [secondaryGoal, setSecondaryGoal] = useState('');
  const [availability, setAvailability] =
    useState<Availability[]>(initialAvailability);
  const [cyclingContext, setCyclingContext] = useState<CyclingContext>(
    initialCyclingContext,
  );
  const [completed, setCompleted] = useState(false);
  const [verificationMessage, setVerificationMessage] = useState('');
  const [sendingVerification, setSendingVerification] = useState(false);
  const [questionnaire, setQuestionnaire] = useState<Questionnaire | null>(
    null,
  );

  useEffect(() => {
    async function load() {
      try {
        const [account, profileResult, questionnaireResult] = await Promise.all(
          [
            apiRequest<{ user: User }>('/v1/me'),
            apiRequest<{ profile: Profile | null }>('/v1/profile'),
            apiRequest<{ questionnaire: Questionnaire }>(
              '/v1/onboarding/questionnaire',
            ),
          ],
        );
        setUser(account.user);
        setQuestionnaire(questionnaireResult.questionnaire);
        if (!profileResult.profile) return;

        const savedProfile = profileResult.profile;
        setProfile({
          birth_date: savedProfile.birth_date || '',
          sex: savedProfile.sex || '',
          height_cm: savedProfile.height_cm?.toString() || '',
          weight_kg: savedProfile.weight_kg?.toString() || '',
          waist_cm: savedProfile.waist_cm?.toString() || '',
          body_fat_percent: savedProfile.body_fat_percent?.toString() || '',
          weight_trend: savedProfile.weight_trend || 'not_informed',
          experience_level: savedProfile.experience_level,
          activity_level: savedProfile.activity_level || '',
        });
        setStep(2);

        const { onboarding } = await apiRequest<{ onboarding: Onboarding }>(
          '/v1/onboarding',
        );
        const savedCyclingContext = onboarding.cycling_context || {};
        setCyclingContext({
          ...initialCyclingContext,
          ...savedCyclingContext,
          discipline: EXCLUDED_DISCIPLINES.has(savedCyclingContext.discipline)
            ? ''
            : savedCyclingContext.discipline || '',
          training_status: TRAINING_STATUSES.has(
            savedCyclingContext.training_status,
          )
            ? savedCyclingContext.training_status
            : 'not_informed',
          preferred_session_types: Array.isArray(
            savedCyclingContext.preferred_session_types,
          )
            ? savedCyclingContext.preferred_session_types
            : [],
        });
        if (onboarding.limitations.length) {
          const saved = onboarding.limitations[0];
          setHasLimitation(true);
          setLimitation({
            kind: saved.kind,
            description: saved.description,
            location: saved.location || '',
            intensity: saved.intensity?.toString() || '',
            aggravating_movement: saved.aggravating_movement || '',
            started_on: saved.started_on || '',
            symptoms_during_after: Array.isArray(saved.symptoms_during_after)
              ? saved.symptoms_during_after
              : [],
            medical_restriction: saved.medical_restriction ?? false,
            recent_surgery: saved.recent_surgery ?? false,
            exercise_prohibited: saved.exercise_prohibited ?? false,
            condition_affecting_exercise:
              saved.condition_affecting_exercise ?? false,
            professional_clearance_recommended:
              saved.professional_clearance_recommended,
          });
        }
        if (onboarding.goals.length) {
          const primary =
            onboarding.goals.find((goal) => goal.priority === 1) ||
            onboarding.goals[0];
          const secondary = onboarding.goals.find(
            (goal) => goal.priority === 2,
          );
          setPrimaryGoal({
            goal_type: primary.goal_type,
            target_date: primary.target_date || '',
            details: primary.details || '',
          });
          setSecondaryGoal(secondary?.goal_type || '');
          setStep(4);
        }
        if (onboarding.availability.length === 7) {
          setAvailability(onboarding.availability);
          setStep(4);
          setCompleted(
            onboarding.availability.some((day) => day.available_minutes > 0),
          );
        }
      } catch (caught) {
        if (isUnauthorized(caught)) {
          window.location.href = '/entrar';
          return;
        }
        setError(
          apiErrorMessage(
            caught,
            'Não foi possível carregar seu perfil. Verifique se a API está em execução e tente novamente.',
          ),
        );
      } finally {
        setLoading(false);
      }
    }
    void load();
  }, []);

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }, [step]);

  // Trocar de etapa limpa os avisos da etapa anterior.
  function goToStep(next: number) {
    setStep(next);
    setError('');
    setMessage('');
  }

  const totalMinutes = useMemo(
    () => availability.reduce((sum, day) => sum + day.available_minutes, 0),
    [availability],
  );
  const trainingDays = useMemo(
    () => availability.filter((day) => day.available_minutes > 0).length,
    [availability],
  );
  const staticCopy = stepCopy[step - 1];
  const contractStep = questionnaire?.steps[step - 1];
  const copy = {
    ...staticCopy,
    title: contractStep?.title || staticCopy.title,
    description: contractStep?.description || staticCopy.description,
  };
  const AsideIcon = copy.icon;

  function updateProfile(field: keyof ProfileForm, value: string) {
    setProfile((current) => ({ ...current, [field]: value }));
  }

  function updateDay(weekday: number, patch: Partial<Availability>) {
    setAvailability((current) =>
      current.map((day) =>
        day.weekday === weekday ? { ...day, ...patch } : day,
      ),
    );
  }

  function toggleSessionPreference(value: string) {
    setCyclingContext((current) => ({
      ...current,
      preferred_session_types: (current.preferred_session_types || []).includes(
        value,
      )
        ? (current.preferred_session_types || []).filter(
            (item) => item !== value,
          )
        : [...(current.preferred_session_types || []), value],
    }));
  }

  function toggleSafetySymptom(value: string) {
    setLimitation((current) => ({
      ...current,
      symptoms_during_after: current.symptoms_during_after.includes(value)
        ? current.symptoms_during_after.filter((item) => item !== value)
        : [...current.symptoms_during_after, value],
    }));
  }

  async function saveProfile(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    await runSave(async () => {
      const result = await apiRequest<{ profile: Profile }>('/v1/profile', {
        method: 'PUT',
        body: JSON.stringify({
          birth_date: profile.birth_date || null,
          sex: profile.sex || null,
          height_cm: profile.height_cm ? Number(profile.height_cm) : null,
          weight_kg: profile.weight_kg ? Number(profile.weight_kg) : null,
          waist_cm: profile.waist_cm ? Number(profile.waist_cm) : null,
          body_fat_percent: profile.body_fat_percent
            ? Number(profile.body_fat_percent)
            : null,
          weight_trend: profile.weight_trend || 'not_informed',
          experience_level: profile.experience_level,
          activity_level: profile.activity_level || null,
        }),
      });
      setProfile((current) => ({
        ...current,
        experience_level: result.profile.experience_level,
      }));
      goToStep(2);
    });
  }

  async function saveLimitations(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    await runSave(async () => {
      const limitations: Limitation[] = hasLimitation
        ? [
            {
              ...limitation,
              intensity: limitation.intensity
                ? Number(limitation.intensity)
                : null,
              started_on: limitation.started_on || null,
              is_active: true,
            },
          ]
        : [];
      await apiRequest('/v1/onboarding/limitations', {
        method: 'PUT',
        body: JSON.stringify({ limitations }),
      });
      goToStep(3);
    });
  }

  async function saveGoals(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    await runSave(async () => {
      const goals: Goal[] = [
        {
          goal_type: primaryGoal.goal_type,
          priority: 1,
          target_date: primaryGoal.target_date || null,
          details: primaryGoal.details,
        },
      ];
      if (secondaryGoal)
        goals.push({
          goal_type: secondaryGoal,
          priority: 2,
          target_date: null,
          details: '',
        });
      await apiRequest('/v1/onboarding/goals', {
        method: 'PUT',
        body: JSON.stringify({ goals }),
      });
      goToStep(4);
    });
  }

  async function saveAvailability(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const updatingCompletedProfile = completed;
    await runSave(async () => {
      await apiRequest('/v1/onboarding/cycling-context', {
        method: 'PUT',
        body: JSON.stringify({ cycling_context: cyclingContext }),
      });
      await apiRequest('/v1/onboarding/availability', {
        method: 'PUT',
        body: JSON.stringify({ availability }),
      });
      setCompleted(true);
      setMessage(
        updatingCompletedProfile
          ? 'Disponibilidade salva. Abra Meu plano e escolha Atualizar plano para gerar um rascunho com esta nova rotina.'
          : 'Perfil concluído. Seu contexto inicial está salvo com segurança.',
      );
    });
  }

  async function runSave(action: () => Promise<void>) {
    setSaving(true);
    setError('');
    setMessage('');
    try {
      await action();
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : 'Não foi possível salvar esta etapa.',
      );
    } finally {
      setSaving(false);
    }
  }

  async function resendVerification() {
    setSendingVerification(true);
    setVerificationMessage('');
    try {
      const result = await apiRequest<{
        message: string;
        development_verification_url?: string;
      }>('/v1/auth/resend-verification', { method: 'POST' });
      setVerificationMessage(
        result.development_verification_url
          ? `${result.message} Abra o link local abaixo.`
          : result.message,
      );
      if (result.development_verification_url)
        window.open(
          result.development_verification_url,
          '_blank',
          'noopener,noreferrer',
        );
    } catch (caught) {
      setVerificationMessage(
        caught instanceof Error
          ? caught.message
          : 'Não foi possível reenviar a confirmação.',
      );
    } finally {
      setSendingVerification(false);
    }
  }

  return {
    AsideIcon,
    availability,
    completed,
    contractStep,
    copy,
    cyclingContext,
    error,
    goToStep,
    hasLimitation,
    limitation,
    loading,
    message,
    primaryGoal,
    profile,
    questionnaire,
    resendVerification,
    runSave,
    saveAvailability,
    saveGoals,
    saveLimitations,
    saveProfile,
    saving,
    secondaryGoal,
    sendingVerification,
    setAvailability,
    setCompleted,
    setCyclingContext,
    setError,
    setHasLimitation,
    setLimitation,
    setLoading,
    setMessage,
    setPrimaryGoal,
    setProfile,
    setQuestionnaire,
    setSaving,
    setSecondaryGoal,
    setSendingVerification,
    setStep,
    setUser,
    setVerificationMessage,
    staticCopy,
    step,
    toggleSafetySymptom,
    toggleSessionPreference,
    totalMinutes,
    trainingDays,
    updateDay,
    updateProfile,
    user,
    verificationMessage,
  };
}

export type ProfileFormController = ReturnType<typeof useProfileForm>;
