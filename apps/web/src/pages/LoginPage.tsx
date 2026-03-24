import SignInForm from "../components/ui-kit/SignInForm";

interface LoginPageProps {
  onLogin: () => void;
}

export default function LoginPage({ onLogin }: LoginPageProps) {
  return <SignInForm onSuccess={onLogin} />;
}
